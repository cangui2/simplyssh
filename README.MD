# SimplySSH

## Description

SimplySSH est une bibliothèque Go conçue pour faciliter la connexion SSH, le téléchargement de fichiers, et la gestion des hôtes à partir d'un fichier de configuration. Elle inclut des fonctionnalités telles que :

- Connexion à un serveur SSH.
- Téléchargement de fichiers et de répertoires via SFTP.
- Lecture et extraction des configurations des hôtes à partir d'un fichier.

SimplySSH is a Go library designed to simplify SSH connections, file downloads, and host management from a configuration file. Features include:

- Connecting to an SSH server.
- Downloading files and directories via SFTP.
- Reading and extracting host configurations from a file.

---

## Installation

### Prérequis / Prerequisites

- [Go](https://golang.org/) (1.16 ou plus récent / 1.16 or newer)
- Une clé SSH valide pour l'authentification / A valid SSH key for authentication

### Instructions

1. Clonez le dépôt / Clone the repository:
   ```bash
   git clone <url_du_dépôt / repository_url>
   cd simplyssh
   ```
2. Importez la bibliothèque dans votre projet Go / Import the library into your Go project:
   ```go
   import "<path_to_package>/simplyssh"
   ```

---

## Utilisation / Usage

### Exemple de configuration des hôtes / Example Host Configuration

Un fichier de configuration SSH typique : / A typical SSH configuration file:

```plaintext
Host example
  HostName example.com
  Port 22
  User user1

Host myserver
  HostName 192.168.1.1
  Port 2222
  User admin
```

### Chargement des hôtes / Loading Hosts

Utilisez la fonction `GetHost` pour charger et lire la configuration des hôtes. / Use the `GetHost` function to load and read host configurations.

```go
hosts := simplyssh.GetHost("path/to/config/file")
for _, host := range hosts {
    fmt.Printf("ID: %d, Host: %s, HostName: %s, Port: %s, User: %s\n",
        host.Id, host.Host, host.HostName, host.Port, host.User)
}
```

### Connexion SSH / SSH Connection

Connectez-vous à un serveur en utilisant la clé SSH. / Connect to a server using an SSH key.

```go
client := simplyssh.Connect("/path/to/private/key", "example.com", "22", "user1")
if client != nil {
    defer client.Close()
}
```

### Téléchargement de fichiers / Downloading Files

#### Téléchargement d'un fichier unique / Download a Single File

```go
err := simplyssh.DownloadFile(sftpClient, "/remote/path/to/file.txt", "/local/path/to/file.txt")
if err != nil {
    log.Fatal(err)
}
```

#### Téléchargement d'un répertoire entier / Download an Entire Directory

```go
err := simplyssh.DownloadDirectory(sftpClient, "/remote/path/to/dir", "/local/path/to/dir")
if err != nil {
    log.Fatal(err)
}
```

---

## Structure du Code / Code Structure

### Fonctions Principales / Main Functions

1. **GetHost**: Charge la configuration des hôtes à partir d'un fichier.

   - Entrée: Chemin du fichier de configuration.
   - Sortie: Liste des hôtes.

2. **Connect**: Établit une connexion SSH avec un serveur.

   - Entrées: Clé SSH, nom d'hôte, port, utilisateur.
   - Sortie: Objet `*ssh.Client`.

3. **DownloadFile**: Télécharge un fichier distant via SFTP.

   - Entrées: Client SFTP, chemin du fichier distant, chemin du fichier local.

4. **DownloadDirectory**: Télécharge récursivement un répertoire distant.

   - Entrées: Client SFTP, chemin du répertoire distant, chemin du répertoire local.

---

## Contribution

Les contributions sont les bienvenues ! / Contributions are welcome!

1. Forkez le dépôt / Fork the repository.
2. Créez une branche pour vos modifications / Create a branch for your changes:
   ```bash
   git checkout -b feature/new-feature
   ```
3. Soumettez une pull request / Submit a pull request.

---

## Licence / License

Ce projet est sous licence MIT. / This project is licensed under the MIT License.

---


## Auteurs / Authors

- Développeur principal : Canguilieme julien
