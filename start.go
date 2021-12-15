package main

import (
	"encoding/json"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var print = fmt.Println

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

type Server struct {
	Id, Alias, User, Ip, PrivateKey, Port string
}

func startCreateServer() {
	id := getUserInput("sample", "id")
	alias := getUserInput("sample", "alias")
	user := getUserInput("sample", "user")
	ip := getUserInput("sample", "ip")
	key := getUserInput("sample", "key")
	port := getUserInput("sample", "port")
	server := newServer(id, alias, user, ip, key, port)

	addServer(server, getServers())
}

func getServerById(id string) Server {
	for _, i := range getServers() {
		if i.Id == id {
			return i
		}
	}
	return newServer("", "", "", "", "", "")
}

func startDeleteServer() {
	listServers()
	id := getUserInput("sample", "id")
	server := getServerById(id)
	removeServer(server, []Server{})
	print("server " + server.Alias + " has been deleted")
}

func startEditServer() {
	listServers()
	id := getUserInput("sample", "id")
	server := getServerById(id)

	alias := getUserInput(server.Alias, "alias")
	user := getUserInput(server.User, "user")
	ip := getUserInput(server.Ip, "ip")
	key := getUserInput(server.PrivateKey, "key")
	port := getUserInput(server.Port, "port")

	server.Alias = alias
	server.User = user
	server.Ip = ip
	server.PrivateKey = key
	server.Port = port

	editServer(server, []Server{})
	print("server " + server.Alias + " has been edited")

}

func newServer(id, alias, user, ip, privateKey, port string) Server {
	return Server{id, alias, user, ip, privateKey, port}
}

func getServers() (servers []Server) {

	fileBytes, err := ioutil.ReadFile("servers.json")

	panicIfError(err)

	err = json.Unmarshal(fileBytes, &servers)
	panicIfError(err)

	return servers
}

func saveServers(servers []Server) {

	videoBytes, err := json.Marshal(servers)
	panicIfError(err)

	err = ioutil.WriteFile("servers.json", videoBytes, 0644)
	panicIfError(err)

}

func addServer(server Server, servers []Server) {
	servers = append(servers, server)
	saveServers(servers)
}

func removeServer(server Server, newServers []Server) {

	currentServers := getServers()

	for _, i := range currentServers {
		if i.Alias == server.Alias {
			continue
		}
		newServers = append(newServers, i)
	}
	saveServers(newServers)

}

func editServer(server Server, newServers []Server) {

	currentServers := getServers()

	for _, i := range currentServers {
		if i.Alias == server.Alias {
			i = server
		}
		newServers = append(newServers, i)
	}
	saveServers(newServers)

}

func listServers() {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("ID", "ALIAS", "USER", "IP", "PRIVATE KEY", "PORT")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, i := range getServers() {
		tbl.AddRow(i.Id, i.Alias, i.User, i.Ip, i.PrivateKey, i.Port)
	}

	tbl.Print()
	print("----------------")
}

func getCommand(service string, server Server) string {
	if service == "ssh" {
		return service + " -i " + server.PrivateKey + " -p " + server.Port + " " + server.User + "@" + server.Ip
	}
	return service + " -i " + server.PrivateKey + " " + server.User + "@" + server.Ip
}

func getServer(id, service string) bool {
	if service != "ssh" && service != "sftp" {
		print("Invalid command: " + service)
		return false
	}

	for _, i := range getServers() {
		if i.Id == id {
			command := getCommand(service, i)
			clipboard.WriteAll(command)
			print("The " + strings.ToUpper(service) + " command for " + "'" + i.Alias + "'" + " is in your clipboard")
			return true
		}
	}
	print("No server found for id " + id)
	return false

}

func createKey(filePath, fileName string) {
	encryptionAlgorithm := "ed25519"
	cmdTwo := exec.Command("ssh-keygen", "-a", "100", "-t", encryptionAlgorithm, "-f", filePath, "-C", fileName)
	cmdThree := exec.Command("ssh-add", filePath)

	_, errTwo := cmdTwo.Output()
	_, errThree := cmdThree.Output()

	panicIfError(errTwo)
	panicIfError(errThree)

	print("New SSH key created:\nPath: " + filePath + "\nEncryption algorithm: " + encryptionAlgorithm)

}

func getUserInput(defaultValue, message string) string {
	ui := defaultValue
	print(message + ": ")
	fmt.Scanln(&ui)
	return ui
}

func getCurrentDirPlusFile(fileName string) string {
	cwd, err := os.Getwd()
	panicIfError(err)
	return filepath.Join(cwd, fileName)
}

func startSSH() {
	listServers()
	serverID := getUserInput("", "Server ID for SSH connection")
	getServer(serverID, "ssh")
}

func startSFTP() {
	listServers()
	serverID := getUserInput("", "Server ID for SFTP connection")
	getServer(serverID, "sftp")
}

func startCreateSSHKey() {
	name := getUserInput("sample", "SSH key name")
	createKey(getCurrentDirPlusFile(name), name)
}

func stringInSlice(value string, list []string) bool {
	for _, v := range list {
		if value == v {
			return true
		}
	}
	return false
}

func listServices() {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("ID", "SERVICE", "ACTION", "DESCRIPTION")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	tbl.AddRow("0", "SSH", "New", "Starts a new SSH connection. Access a remote server (default option)")
	tbl.AddRow("1", "SFTP", "New", "Starts a new SFTP connection. Transfer files back and forth between your computer and the remote server")
	tbl.AddRow("2", "SSH KEY", "Create", "Creates a SSH key (public and private). Start SSH/SFTP connections without passwords")
	tbl.AddRow("3", "Server","Add", "Adds a new server to your servers list")
	tbl.AddRow("4", "Server","Delete", "Deletes a server from your servers list")
	tbl.AddRow("5", "Server", "Edit","Edits a server from your servers list")

	tbl.Print()
	print("----------------")
}

func startService() {
	// 	text := `
	// Select a service:
	// 0-SSH -> Start connection (default)
	// 1-SFTP -> Start connection
	// 2-SSH key -> New
	// 3-Server -> New
	// 4-Server -> Detete
	// 5-Server -> Edit`
	listServices()
	service := getUserInput("0", "Service ID")

	servicesChoices := []string{"0", "1", "2", "3", "4", "5"}

	if !stringInSlice(service, servicesChoices) {
		print("Invalid service: " + service)
		return
	}

	choices := map[string]interface{}{}

	choices["0"] = startSSH
	choices["1"] = startSFTP
	choices["2"] = startCreateSSHKey
	choices["3"] = startCreateServer
	choices["4"] = startDeleteServer
	choices["5"] = startEditServer

	choices[service].(func())()

}

func main() {
	print("---- SSH/SFTP MANAGER ----")
	// listServers()
	// listServices()
	startService()

}
