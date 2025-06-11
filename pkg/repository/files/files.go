package files

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"telegrambot/internal/e"
	"telegrambot/pkg/repository"
	"time"
)

type RepositoryFiles struct {
	basePath string
}

func New(basePath string) RepositoryFiles {
	return RepositoryFiles{basePath: basePath}
}

const defaultPerm = 0774

func (r RepositoryFiles) Save(ctx context.Context, page *repository.Page) (err error) {
	defer func() {
		err = e.WrapIfErr("can't save page", err)
	}()

	fPath := filepath.Join(r.basePath, page.Username)

	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	fName, err := fileName(page)
	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fName)

	file, err := os.Create(fPath)
	if err != nil {
		return err
	}
	defer func() {
		err = e.WrapIfErr("can't create", file.Close())
	}()

	if err = gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (r RepositoryFiles) PickRandom(ctx context.Context, username string) (page *repository.Page, err error) {
	defer func() {
		err = e.WrapIfErr("can't pick random page", err)
	}()

	fPath := filepath.Join(r.basePath, username)

	files, err := os.ReadDir(fPath)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, repository.ErrNoSavedPages
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	n := rand.Intn(len(files))

	file := files[n]

	return r.decodePage(filepath.Join(fPath, file.Name()))
}

func (r RepositoryFiles) Remove(ctx context.Context, p *repository.Page) error {
	fName, err := fileName(p)
	if err != nil {
		return e.WrapIfErr("can't remove file", err)
	}

	fPath := filepath.Join(r.basePath, p.Username, fName)

	if err := os.Remove(fPath); err != nil {
		msg := fmt.Sprintf("can't remove file: %s", fPath)
		return e.WrapIfErr(msg, err)
	}

	return nil
}

func (r RepositoryFiles) IsExists(ctx context.Context, p *repository.Page) (bool, error) {
	fName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file exists", err)
	}

	fPath := filepath.Join(r.basePath, p.Username, fName)

	switch _, err = os.Stat(fPath); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't check if file %s exists", fPath)
		return false, e.Wrap(msg, err)
	}

	return true, nil
}

func (r RepositoryFiles) decodePage(filePath string) (page *repository.Page, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}
	defer func() {
		err = e.WrapIfErr("can't close file", file.Close())
	}()

	var p repository.Page

	if err := gob.NewDecoder(file).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

func fileName(p *repository.Page) (string, error) {
	return p.Hash()
}
