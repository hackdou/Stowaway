package agent

import (
	"Stowaway/common"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

var (
	Stdin   io.Writer
	Stdout  io.Reader
	Sshhost *ssh.Session
)

//启动ssh
func StartSSH(info string, nodeid string) error {
	var authpayload ssh.AuthMethod
	spiltedinfo := strings.Split(info, ":::")
	host := spiltedinfo[0]
	username := spiltedinfo[1]
	authway := spiltedinfo[2]
	method := spiltedinfo[3]

	if method == "1" {
		authpayload = ssh.Password(authway)
	} else if method == "2" {
		key, err := ssh.ParsePrivateKey([]byte(authway))
		if err != nil {
			sshMess, _ := common.ConstructPayload(common.AdminId, "", "COMMAND", "SSHCERTERROR", " ", " ", 0, nodeid, AgentStatus.AESKey, false)
			ProxyChan.ProxyChanToUpperNode <- sshMess
			return err
		}
		authpayload = ssh.PublicKeys(key)
	}

	sshdial, err := ssh.Dial("tcp", host, &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{authpayload},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		sshMess, _ := common.ConstructPayload(common.AdminId, "", "COMMAND", "SSHRESP", " ", "FAILED", 0, nodeid, AgentStatus.AESKey, false)
		ProxyChan.ProxyChanToUpperNode <- sshMess
		return err
	}
	Sshhost, err = sshdial.NewSession()

	if err != nil {
		sshMess, _ := common.ConstructPayload(common.AdminId, "", "COMMAND", "SSHRESP", " ", "FAILED", 0, nodeid, AgentStatus.AESKey, false)
		ProxyChan.ProxyChanToUpperNode <- sshMess
		return err
	}
	Stdout, err = Sshhost.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}
	Stdin, err = Sshhost.StdinPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}
	Sshhost.Stderr = Sshhost.Stdout
	Sshhost.Shell()
	sshMess, _ := common.ConstructPayload(common.AdminId, "", "COMMAND", "SSHRESP", " ", "SUCCESS", 0, nodeid, AgentStatus.AESKey, false)
	ProxyChan.ProxyChanToUpperNode <- sshMess
	return nil
}

func WriteCommand(command string) {
	Stdin.Write([]byte(command))
}

func ReadCommand() {
	buffer := make([]byte, 10240)
	for {
		len, err := Stdout.Read(buffer)
		if err != nil {
			break
		}
		sshRespMess, _ := common.ConstructPayload(common.AdminId, "", "DATA", "SSHMESS", " ", string(buffer[:len]), 0, AgentStatus.Nodeid, AgentStatus.AESKey, false)
		ProxyChan.ProxyChanToUpperNode <- sshRespMess
	}
}
