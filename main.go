package simplyssh

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type hostConnection struct {
	Id       uint
	Host     string
	HostName string
	Port     string
	User     string
}
func GetHost(path string) []hostConnection {
	fileData, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil
	}

	allHosts := []hostConnection{}

	word := []byte{}
	breakLine := "\n"
	//chargement initiale de lobject
	var currentHost hostConnection
	var i int=0
	for _, data := range fileData {
		if !bytes.Equal([]byte{data}, []byte(breakLine)) {
			word = append(word, data)
		} else {
			line := strings.TrimSpace(string(word))
			if strings.HasPrefix(line, "Host ") {
				// Si un nouveau Host est trouvé, ajouter l'ancien à la liste
				if currentHost.Host != "" {
					allHosts = append(allHosts, currentHost)
				}
				// Initialiser un nouveau hostConnection
				currentHost = hostConnection{}
				currentHost.Id=uint(i)
				parts := strings.Fields(line)
				if len(parts) > 1 {
					currentHost.Host = parts[1]
				}
			} else if strings.HasPrefix(line, "HostName ") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					currentHost.HostName = parts[1]
				}
			} else if strings.HasPrefix(line, "Port ") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					currentHost.Port = parts[1]
				}
			} else if strings.HasPrefix(line, "User ") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					currentHost.User = parts[1]
					i++
				}
			}
			word = word[:0]
		}
	
	}

	// Ajouter le dernier host traité
	if currentHost.Host != "" {
		allHosts = append(allHosts, currentHost)
	}

	return allHosts
}

func Connect(pathKey, hostName, port, user string) *ssh.Client {

	sshConfig := sshConfig(user, pathKey)
	connection, err := ssh.Dial("tcp", hostName+":"+port, sshConfig)
	if err != nil {
        log.Fatalf("Failed to dial: %s", err) 
		return nil
	}
	fmt.Println("Connection ok ")

	return connection
}

func sshConfig(usernamme, path string) *ssh.ClientConfig {
	sshConfig := &ssh.ClientConfig{
		User: usernamme,
		Auth: []ssh.AuthMethod{
			PublicKeyFile(path),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Insecure, use for testing only


		
	}
	return sshConfig

}
func PublicKeyFile(file string) ssh.AuthMethod {
	fmt.Println(file)

	buffer, err := os.ReadFile(file)
	if err != nil {
		fmt.Errorf("Failed to buffer: %s", err)

		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		fmt.Errorf("Failed to key: %s", err)

		return nil
	}
	return ssh.PublicKeys(key)
}
type ProgressReader struct {
    reader     io.Reader
    totalSize  int64
    bytesRead  int64
    lastUpdate time.Time
}

func NewProgressReader(r io.Reader, totalSize int64) *ProgressReader {
    return &ProgressReader{
        reader:    r,
        totalSize: totalSize,
    }
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
    n, err := pr.reader.Read(p)
    pr.bytesRead += int64(n)

    // Mettre à jour la progression toutes les 500 ms par exemple
    now := time.Now()
    if now.Sub(pr.lastUpdate) > 500*time.Millisecond || err == io.EOF {
        pr.lastUpdate = now
        if pr.totalSize > 0 {
            percent := float64(pr.bytesRead) / float64(pr.totalSize) * 100
            fmt.Printf("\r%.2f%% téléchargé...", percent)
        } else {
            // Si on ne connaît pas la taille, on ne peut afficher que les octets
            fmt.Printf("\r%v bytes téléchargés...", pr.bytesRead)
        }
    }

    return n, err
}

func DownloadDirectory(sftpClient *sftp.Client, remoteDir, localDir string) error {
    entries, err := sftpClient.ReadDir(remoteDir)
    if err != nil {
        return err
    }

    // Crée le répertoire local s’il n’existe pas
    os.MkdirAll(localDir, 0755)

    for _, entry := range entries {
        remotePath := remoteDir + "/" + entry.Name()
        localPath := localDir + "/" + entry.Name()

        if entry.IsDir() {
            // Téléchargement récursif
            err = DownloadDirectory(sftpClient, remotePath, localPath)
            if err != nil {
                return err
            }
        } else {
            // Téléchargement de fichier
            remoteFile, err := sftpClient.Open(remotePath)
            if err != nil {
                return err
            }
            defer remoteFile.Close()

            localFile, err := os.Create(localPath)
            if err != nil {
                return err
            }
            defer localFile.Close()

            // Récupère la taille du fichier distant
            info, err := remoteFile.Stat()
            if err != nil {
                return err
            }

            totalSize := info.Size()

            // Crée un ProgressReader qui va afficher la progression
            progressReader := NewProgressReader(remoteFile, totalSize)

            _, err = io.Copy(localFile, progressReader)
            if err != nil {
                return err
            }
            fmt.Println("\nTéléchargement terminé:", localPath)
        }
    }
    return nil
}
func DownloadFile(sftpClient *sftp.Client, remoteFilePath, localFilePath string) error {
    // Ouvrir le fichier distant
    remoteFile, err := sftpClient.Open(remoteFilePath)
    if err != nil {
        return err
    }
    defer remoteFile.Close()

    // Créer le répertoire local si nécessaire
    localDir := filepath.Dir(localFilePath)
    if err := os.MkdirAll(localDir, 0755); err != nil {
        return err
    }

    // Créer le fichier local ou ecrase le dossier
    localFile, err := os.Create(localFilePath)
    if err != nil {
        return err
    }
    defer localFile.Close()

    // Récupère la taille du fichier distant
    info, err := remoteFile.Stat()
    if err != nil {
        return err
    }
    totalSize := info.Size()

    // Crée un ProgressReader qui va afficher la progression
    progressReader := NewProgressReader(remoteFile, totalSize)

    // Copie du fichier avec affichage de la progression
    _, err = io.Copy(localFile, progressReader)
    if err != nil {
        return err
    }

    fmt.Println("\nTéléchargement terminé:", localFilePath)
    return nil
}
