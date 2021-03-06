package utils

import (
  "crypto/x509"
  "encoding/pem"
  "golang.org/x/crypto/ssh"
  //"golang.org/x/crypto/ssh"
  //"io"
  "io/ioutil"
  //"os"
  "log"
  "bytes"
  "fmt"
)

func GetConfig(user string, pwd string, key string) ssh.ClientConfig {
  pemKey, err := ioutil.ReadFile(key)
  if err != nil {
      log.Fatalf("unable to read private key: %v", err)
  }

  block, _ := pem.Decode([]byte(pemKey))

  derKey, err := x509.DecryptPEMBlock(block, []byte(pwd))
  if err != nil {
      log.Fatalf("unable to decrypt private key: %v", err)
  }

  privKey, err := x509.ParsePKCS1PrivateKey(derKey)
  if err != nil {
      log.Fatalf("unable to decrypt pkcs1 private key: %v", err)
  }

  signer, err := ssh.NewSignerFromKey(privKey)
  if err != nil {
      log.Fatalf("unable to parse private key: %v", err)
  }

  return ssh.ClientConfig{
  	User: user,
  	Auth: []ssh.AuthMethod{
      ssh.PublicKeys(signer),
    },
  }
}

func CreateConnection(config ssh.ClientConfig, server string) *ssh.Client {
  conn, err := ssh.Dial("tcp", server, &config)
  if err != nil {
    log.Fatalf("unable to connect: %s", err)
  }
  return conn
}

func RunCommand(conn *ssh.Client, command string) string {
  session, err := conn.NewSession()
  if err != nil {
    log.Fatalf("unable to create session: %s", err)
  }
  defer session.Close()

  var stdoutBuf bytes.Buffer
  session.Stdout = &stdoutBuf
  err = session.Run(command)
  if err != nil {
    log.Fatalf("Failed to get listing: %s", err)
  }
  return stdoutBuf.String()
}

func CopyToServer(conn *ssh.Client, filename string, content string) {
  session, err := conn.NewSession()
  if err != nil {
    log.Fatalf("unable to create session: %s", err)
  }
  defer session.Close()

  go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		//fmt.Fprintln(w, "D0755", 0, "testdir") // mkdir
		fmt.Fprintln(w, "C0644", len(content), filename)
		fmt.Fprint(w, content)
		fmt.Fprint(w, "\x00") // transfer end with \x00
	}()
	if err := session.Run("/usr/bin/scp -tr ./"); err != nil {
    log.Fatalf("unable to scp: %s", err)
	}
}
