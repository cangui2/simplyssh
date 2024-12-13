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

// GetHost reads the SSH configuration file and extracts host connections.
// GetHost lit le fichier de configuration SSH et extrait les connexions des hôtes.
func GetHost(path string) []hostConnection {
	fileData, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", err) // Erreur lors de la lecture du fichier.
		return nil
	}

	allHosts := []hostConnection{}

	word := []byte{}
	breakLine := "\n"
	// Initial object loading / Chargement initial de l'objet
	var currentHost hostConnection
	var i int = 0
	for _, data := range fileData {
		if !bytes.Equal([]byte{data}, []byte(breakLine)) {
			word = append(word, data)
		} else {
			line := strings.TrimSpace(string(word))
			if strings.HasPrefix(line, "Host ") {
				// If a new Host is found, add the previous one to the list
				// Si un nouvel hôte est trouvé, ajouter le précédent à la liste
				if currentHost.Host != "" {
					allHosts = append(allHosts, currentHost)
				}
				// Initialize a new hostConnection / Initialiser une nouvelle connexion hôte
				currentHost = hostConnection{}
				currentHost.Id = uint(i)
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

	// Add the last processed host / Ajouter le dernier hôte traité
	if currentHost.Host != "" {
		allHosts = append(allHosts, currentHost)
	}

	return allHosts
}

// Connect establishes an SSH connection to the remote host.
// Connect établit une connexion SSH à l'hôte distant.
func Connect(pathKey, hostName, port, user string) *ssh.Client {
	sshConfig := sshConfig(user, pathKey)
	connection, err := ssh.Dial("tcp", hostName+":"+port, sshConfig)
	if err != nil {
		log.Fatalf("Failed to dial: %s", err) // Échec de la connexion
		return nil
	}
	fmt.Println("Connection ok ") // Connexion réussie

	return connection
}

// sshConfig generates the SSH client configuration.
// sshConfig génère la configuration du client SSH.
func sshConfig(usernamme, path string) *ssh.ClientConfig {
	sshConfig := &ssh.ClientConfig{
		User: usernamme,
		Auth: []ssh.AuthMethod{
			PublicKeyFile(path),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Insecure, use for testing only / Insecure, à utiliser uniquement pour les tests
	}
	return sshConfig
}

// PublicKeyFile reads the private key from a file and returns an SSH AuthMethod.
// PublicKeyFile lit la clé privée depuis un fichier et retourne une méthode d'authentification SSH.
func PublicKeyFile(file string) ssh.AuthMethod {
	fmt.Println(file)

	buffer, err := os.ReadFile(file)
	if err != nil {
		fmt.Errorf("Failed to buffer: %s", err) // Échec du chargement du buffer
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		fmt.Errorf("Failed to key: %s", err) // Échec du parsing de la clé
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

// NewProgressReader creates a reader to display progress during file transfers.
// NewProgressReader crée un lecteur pour afficher la progression lors des transferts de fichiers.
func NewProgressReader(r io.Reader, totalSize int64) *ProgressReader {
	return &ProgressReader{
		reader:    r,
		totalSize: totalSize,
	}
}

// Read reads data and updates the progress display.
// Read lit les données et met à jour l'affichage de la progression.
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.bytesRead += int64(n)

	// Update progress every 500 ms / Mettre à jour la progression toutes les 500 ms
	now := time.Now()
	if now.Sub(pr.lastUpdate) > 500*time.Millisecond || err == io.EOF {
		pr.lastUpdate = now
		if pr.totalSize > 0 {
			percent := float64(pr.bytesRead) / float64(pr.totalSize) * 100
			fmt.Printf("\r%.2f%% downloaded...", percent) // téléchargé...
		} else {
			fmt.Printf("\r%v bytes downloaded...", pr.bytesRead) // octets téléchargés
		}
	}

	return n, err
}

// DownloadDirectory downloads a remote directory recursively to the local filesystem.
// DownloadDirectory télécharge un répertoire distant de manière récursive vers le système de fichiers local.
func DownloadDirectory(sftpClient *sftp.Client, remoteDir, localDir string) error {
	entries, err := sftpClient.ReadDir(remoteDir)
	if err != nil {
		return err
	}

	// Create the local directory if it does not exist / Crée le répertoire local s’il n’existe pas
	os.MkdirAll(localDir, 0755)

	for _, entry := range entries {
		remotePath := remoteDir + "/" + entry.Name()
		localPath := localDir + "/" + entry.Name()

		if entry.IsDir() {
			// Recursive download / Téléchargement récursif
			err = DownloadDirectory(sftpClient, remotePath, localPath)
			if err != nil {
				return err
			}
		} else {
			// File download / Téléchargement de fichier
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

			// Get the remote file size / Récupère la taille du fichier distant
			info, err := remoteFile.Stat()
			if err != nil {
				return err
			}

			totalSize := info.Size()

			// Create a ProgressReader to display the progress / Crée un ProgressReader qui va afficher la progression
			progressReader := NewProgressReader(remoteFile, totalSize)

			_, err = io.Copy(localFile, progressReader)
			if err != nil {
				return err
			}
			fmt.Println("\nDownload completed:", localPath) // Téléchargement terminé
		}
	}
	return nil
}

// DownloadFile downloads a single file from the remote server.
// DownloadFile télécharge un fichier unique depuis le serveur distant.
func DownloadFile(sftpClient *sftp.Client, remoteFilePath, localFilePath string) error {
	// Open the remote file / Ouvrir le fichier distant
	remoteFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	// Create the local directory if needed / Créer le répertoire local si nécessaire
	localDir := filepath.Dir(localFilePath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	// Create or overwrite the local file / Créer ou écraser le fichier local
	localFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// Get the remote file size / Récupère la taille du fichier distant
	info, err := remoteFile.Stat()
	if err != nil {
		return err
	}
	totalSize := info.Size()

	// Create a ProgressReader to display the progress / Crée un ProgressReader pour afficher la progression
	progressReader := NewProgressReader(remoteFile, totalSize)

	// Copy the file with progress display / Copier le fichier avec affichage de la progression
	_, err = io.Copy(localFile, progressReader)
	if err != nil {
		return err
	}

	fmt.Println("\nDownload completed:", localFilePath) // Téléchargement terminé
	return nil
}

// ExecuteCommand runs a command on the remote host via SSH.
// ExecuteCommand exécute une commande sur l'hôte distant via SSH.
func ExecuteCommand(command string, session *ssh.Session) {
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		fmt.Println(b)
		log.Fatal("Failed cmd: ", err) // Échec de la commande
	}
	fmt.Println(b)
}
