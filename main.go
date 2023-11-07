package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"pack.ag/tftp"
	"path/filepath"
)

const TftpDir = "/tftpboot"

func main() {
	s, err := tftp.NewServer(":69", tftp.ServerSinglePort(true))
	if err != nil {
		panic(err)
	}
	readHandler := tftp.ReadHandlerFunc(serveTFTP)
	writeHandler := tftp.WriteHandlerFunc(receiveTFTP)
	s.ReadHandler(readHandler)
	s.WriteHandler(writeHandler)
	s.ListenAndServe()
	select {}

}

func errorDefer(fn func() error, msg string) {
	if err := fn(); err != nil {
		log.Printf("[DEBUG] "+msg+": %v\n", err)
	}
}

func serveTFTP(r tftp.ReadRequest) {
	log.Printf("[%s] GET %s\n", r.Addr().IP.String(), r.Name())
	path := filepath.Join(TftpDir, filepath.Clean(r.Name()))

	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		r.WriteError(tftp.ErrCodeFileNotFound, fmt.Sprintf("File %q does not exist", r.Name()))
		return
	}
	defer errorDefer(file.Close, "error closing file")

	finfo, _ := file.Stat()
	r.WriteSize(finfo.Size())
	if _, err = io.Copy(r, file); err != nil {
		log.Println(err)
	}
}

func receiveTFTP(w tftp.WriteRequest) {
	log.Printf("[%s] PUT %s\n", w.Addr().IP.String(), w.Name())
	path := filepath.Join(TftpDir, filepath.Clean(w.Name()))

	file, err := os.Create(path)
	if err != nil {
		log.Println(err)
		w.WriteError(tftp.ErrCodeAccessViolation, fmt.Sprintf("Cannot create file %q", filepath.Clean(w.Name())))
	}
	defer errorDefer(file.Close, "error closing file")

	_, err = io.Copy(file, w)
	if err != nil {
		log.Println(err)
	}
}
