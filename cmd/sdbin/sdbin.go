package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

const (
	headerLength = 32
	hashOffset   = 16
	saltOffset   = 893400
	saltLength   = 8
)

var header = []byte{
	0x41, 0x49, 0x54, 0x53, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // fixed header
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // md5 bytes
}

var md5Salt = []byte{'V', 'D', 'R', 'A', 'C', '9', '9', '9'}

func main() {
	argCount := len(os.Args)
	if argCount < 2 {
		fmt.Println(`This utility creates an encoded file
FW is updated with this file in SD card
			
Command Format:
sdbin source [dest]
			
source:         source is input file name
dest:           dest is output file name
                default is "SD_CarDV.bin"
		`)
		os.Exit(1)
	}

	inputFilename := os.Args[1]
	outputFilename := `SD_CarDV.bin`
	if argCount == 3 {
		outputFilename = os.Args[2]
	}

	if err := sdbin(inputFilename, outputFilename); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func sdbin(inputFilename, outputFilename string) (err error) {
	stat, err := os.Stat(inputFilename)
	if err != nil {
		return fmt.Errorf(`unable to access %q: %w`, inputFilename, err)
	}

	fmt.Printf("Input file size %d\n", stat.Size())

	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return fmt.Errorf(`unable to open %q for reading: %w`, inputFilename, err)
	}
	defer inputFile.Close()

	// Added sanity check, just trying *not* to brick cameras
	sanityCheck(inputFile)

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf(`unable to create %q: %w`, outputFilename, err)
	}
	defer func() {
		cerr := outputFile.Close()
		if cerr != nil {
			err = fmt.Errorf(`something went wrong closing %q: %w`, outputFilename, cerr)
		}

		// Cleanup the output file if things go wrong
		if err != nil {
			os.Remove(outputFilename)
		}
	}()

	return sdbinStreams(inputFile, outputFile, stat.Size())
}

func sdbinStreams(inputFile io.ReadSeeker, outputFile io.ReadWriteSeeker, fileSize int64) error {
	written, err := outputFile.Write(header)
	if err := check(headerLength, written, err); err != nil {
		return fmt.Errorf(`unable to write the header: %w`, err)
	}

	md5Hasher := md5.New()
	written, err = md5Hasher.Write(md5Salt)
	if err := check(len(md5Salt), written, err); err != nil {
		return fmt.Errorf("failed to salt the MD5: %w", err)
	}

	written64, err := io.Copy(io.MultiWriter(outputFile, md5Hasher), inputFile)
	if err := check64(fileSize, written64, err); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	hashsum := md5Hasher.Sum(nil)
	fmt.Printf("%x\n", hashsum)

	if err := seekSet(outputFile, hashOffset); err != nil {
		return fmt.Errorf(`unable to seek hash in the header: %w`, err)
	}

	written, err = outputFile.Write(hashsum)
	if err := check(len(hashsum), written, err); err != nil {
		return fmt.Errorf("failed to write hash to header: %w", err)
	}
	return nil
}

func sanityCheck(r io.ReadSeeker) error {
	err := seekSet(r, saltOffset)
	if err != nil {
		return fmt.Errorf(`sanity check failed due to %w`, err)
	}

	buf := make([]byte, saltLength)
	read, err := r.Read(buf)
	if err := check(saltLength, read, err); err != nil {
		return fmt.Errorf(`tried to read the stored salt but %w`, err)
	}

	if !bytes.Equal(buf, md5Salt) {
		return fmt.Errorf(`sanity check failed, was seeking to find "%0x" but got "%0x" are you sure this is a MiVue firmware file?`, md5Salt, buf)
	}

	if err := seekSet(r, 0); err != nil {
		return fmt.Errorf(`sanity check failed due to %w`, err)
	}

	return nil
}

func seekSet(r io.Seeker, offset int64) error {
	position, err := r.Seek(offset, os.SEEK_SET)
	if err != nil {
		return fmt.Errorf(`unable to seek to %d due to %w`, offset, err)
	}
	if position != offset {
		return fmt.Errorf(`tried to seek to %d landed at %d`, offset, position)
	}
	return nil
}

func check(want, got int, err error) error {
	return check64(int64(want), int64(got), err)
}

func check64(want, got int64, err error) error {
	if err != nil {
		return err
	}
	if want > got {
		return fmt.Errorf(`only managed %d out of %d bytes`, got, want)
	}
	if want < got {
		return fmt.Errorf(`got extra bytes, got %d but wanted %d bytes`, got, want)
	}
	return nil
}
