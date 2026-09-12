package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type matter struct {
	Number      string
	DocumentKey string
	Deadline    time.Time
	SignedReady bool
}

func followUpStatus(m matter, now time.Time) string {
	if !m.SignedReady {
		return "awaiting-signed-document"
	}
	if now.After(m.Deadline) {
		return "deadline-follow-up"
	}
	return "signed-document-delivery"
}

func uploadPart(url string, data []byte) (string, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("part upload returned %s", res.Status)
	}
	return res.Header.Get("ETag"), nil
}

func runMatterIntake(path string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	const bucket = "legal-matter-media"
	if err := c.createBucket(bucket); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	m := matter{Number: "MAT-1042", DocumentKey: "matters/MAT-1042/signed-delivery.bin", Deadline: time.Now().Add(24 * time.Hour)}
	created, err := c.createMultipart(bucket, m.DocumentKey)
	if err != nil {
		return err
	}
	uploadID, ok := created["upload_id"].(string)
	if !ok || uploadID == "" {
		return fmt.Errorf("create response has no upload_id")
	}
	parts := make([]map[string]any, 0)
	buf := make([]byte, 5*1024*1024)
	for partNumber := 1; ; partNumber++ {
		n, readErr := io.ReadFull(f, buf)
		if readErr == io.EOF {
			break
		}
		if readErr != nil && readErr != io.ErrUnexpectedEOF {
			return readErr
		}
		presigned, err := c.presignPart(uploadID, partNumber)
		if err != nil {
			return err
		}
		url, ok := presigned["url"].(string)
		if !ok || url == "" {
			return fmt.Errorf("presign response has no url")
		}
		etag, err := uploadPart(url, buf[:n])
		if err != nil {
			return err
		}
		parts = append(parts, map[string]any{"part_number": partNumber, "etag": etag})
		if readErr == io.ErrUnexpectedEOF {
			break
		}
	}
	if _, err := c.completeMultipart(uploadID, parts); err != nil {
		return err
	}
	m.SignedReady = true
	fmt.Printf("%s: %s (%d parts)\n", m.Number, followUpStatus(m, time.Now()), len(parts))
	return nil
}
