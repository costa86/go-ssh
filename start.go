package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

var print = fmt.Println

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

type Server struct {
	Id, Alias, User, Ip, PrivateKey, Port, Notes string
}

func validateSSHKeyFile(key string) bool {
	_, err := os.Stat(key)
	if err != nil {
		print("SSH key file not found: " + key)
		return false
	}
	pub := key[len(key)-4:]
	if pub == ".pub" {
		print(key + " is the public key. You must use the private part (the one with no extension)")
		return false
	}
	return true
}

func startCreateServer() {
	print("Enter the values for a new server\nDo not add spaces in any field!")
	id := getUserInput("sample", "ID")
	alias := getUserInput("sample", "Alias")
	user := getUserInput("sample", "User")
	ip := getUserInput("sample", "IP")
	key := getUserInput("sample", "SSH key")

	if !validateSSHKeyFile(key) {
		return
	}

	port := getUserInput("sample", "Port")
	notes := getUserInput("sample", "Notes")

	server := newServer(id, alias, user, ip, key, port, notes)

	addServer(server, getServers())
	print("server " + server.Alias + " has been CREATED")

}

func getServerById(id string) Server {
	for _, i := range getServers() {
		if i.Id == id {
			return i
		}
	}
	return newServer("", "", "", "", "", "", "")
}

func startDeleteServer() {
	print("Identify the server you want to DELETE")
	listServers()
	id := getUserInput("sample", "ID")
	server := getServerById(id)
	if server.Id == "" {
		print("Server id " + id + " was not found")
		return
	}
	removeServer(server, []Server{})
	print("server " + server.Alias + " has been DELETED")
}

func startEditServer() {
	print("Identify the server you want to EDIT\nJust hit 'enter' for the values you do not want to change")
	listServers()
	id := getUserInput("sample", "ID")
	server := getServerById(id)

	if server.Id == "" {
		print("Server id " + id + " was not found")
		return
	}
	print("Enter the new values for " + server.Alias + "\nDo not add spaces in any field!")

	alias := getUserInput(server.Alias, "Alias")
	user := getUserInput(server.User, "User")
	ip := getUserInput(server.Ip, "IP")
	key := getUserInput(server.PrivateKey, "SSH key")

	if !validateSSHKeyFile(key) {
		return
	}

	port := getUserInput(server.Port, "Port")
	notes := getUserInput(server.Notes, "Notes")

	server.Alias = alias
	server.User = user
	server.Ip = ip
	server.PrivateKey = key
	server.Port = port
	server.Notes = notes

	editServer(server, []Server{})
	print("server " + server.Alias + " has been EDITED")

}

func newServer(id, alias, user, ip, privateKey, port, notes string) Server {
	return Server{id, alias, user, ip, privateKey, port, notes}
}

func createServerFileAndGetServers() (servers []Server) {
	server := newServer("sample", "sample", "sample", "sample", "sample", "sample", "sample")
	addServer(server, []Server{})
	fileBytes, er := ioutil.ReadFile("servers.json")
	panicIfError(er)

	err := json.Unmarshal(fileBytes, &servers)
	panicIfError(err)
	return servers
}

func getServers() (servers []Server) {

	fileBytes, err := ioutil.ReadFile("servers.json")

	if err != nil {
		return createServerFileAndGetServers()
	}

	err = json.Unmarshal(fileBytes, &servers)
	panicIfError(err)

	return servers
}

func saveServersToFile(servers []Server) {

	videoBytes, err := json.Marshal(servers)
	panicIfError(err)

	err = ioutil.WriteFile("servers.json", videoBytes, 0644)
	panicIfError(err)

}

func addServer(server Server, servers []Server) {
	servers = append(servers, server)
	saveServersToFile(servers)
}

func removeServer(server Server, newServers []Server) {

	for _, i := range getServers() {
		if i.Alias == server.Alias {
			continue
		}
		newServers = append(newServers, i)
	}
	saveServersToFile(newServers)

}

func editServer(server Server, newServers []Server) {

	for _, i := range getServers() {
		if i.Id == server.Id {
			i = server
		}
		newServers = append(newServers, i)
	}
	saveServersToFile(newServers)

}

func listServers() {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("ID", "ALIAS", "USER", "IP", "PRIVATE KEY", "PORT", "NOTES")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, i := range getServers() {
		tbl.AddRow(i.Id, i.Alias, i.User, i.Ip, i.PrivateKey, i.Port, i.Notes)
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

func sendCommandToClipboard(id, service string) bool {

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

	cmdSSHKeyGen := exec.Command("ssh-keygen", "-a", "100", "-t", encryptionAlgorithm, "-f", filePath, "-C", fileName)
	cmdSSHAdd := exec.Command("ssh-add", filePath)

	_, errSSHKeyGen := cmdSSHKeyGen.Output()
	_, errSSHAdd := cmdSSHAdd.Output()

	panicIfError(errSSHKeyGen)
	panicIfError(errSSHAdd)

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
	sendCommandToClipboard(serverID, "ssh")
}

func startSFTP() {
	listServers()
	serverID := getUserInput("", "Server ID for SFTP connection")
	sendCommandToClipboard(serverID, "sftp")
}

func startCreateSSHKey() {
	name := getUserInput("key", "SSH key name")
	path := getCurrentDirPlusFile(name)
	createKey(path, name)
}

func textInList(value string, list []string) bool {
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
	tbl.AddRow("3", "Server", "Add", "Adds a new server to your servers list")
	tbl.AddRow("4", "Server", "Delete", "Deletes a server from your servers list")
	tbl.AddRow("5", "Server", "Edit", "Edits a server from your servers list")

	tbl.Print()
	print("----------------")
}

func startService() {
	listServices()

	selectedServiceID := getUserInput("0", "Service ID")
	servicesChoices := []string{"0", "1", "2", "3", "4", "5"}

	if !textInList(selectedServiceID, servicesChoices) {
		print("Invalid service: " + selectedServiceID)
		return
	}

	service := map[string]interface{}{}

	service["0"] = startSSH
	service["1"] = startSFTP
	service["2"] = startCreateSSHKey
	service["3"] = startCreateServer
	service["4"] = startDeleteServer
	service["5"] = startEditServer

	service[selectedServiceID].(func())()
}

func main() {
	print("---- SSH/SFTP MANAGER ----")
	startService()
}
