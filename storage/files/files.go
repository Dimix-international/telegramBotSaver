package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"telegramBotSaver/lib/e"
	"telegramBotSaver/storage"
)

type Storage struct {
	basePath string //в какой папке будем хранить
}

//permissions
const defaultPerm = 0744 //у всех права на чтение и запись

func New(basePath string) *Storage {
	return &Storage{
		basePath: basePath,
	}
}

func (s *Storage) Save(page *storage.Page) (err error) {
	defer func () {err = e.WrapIfErr("can't save page", err)} ()

	//путь сохранения файла
	fPath := filepath.Join(s.basePath, page.UserName) //filepath.Join будет работать независимо от операц системы

	//создаем директории по пути
	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	//делаем название файла (уникальные), используем хэш
	fName, err := fileName(page)
	if err != nil {
		return err;
	}

	//дописываем путь до имени файла 
	fPath = filepath.Join(fPath, fName)

	//создаем файл
	file, err := os.Create(fPath)
	if err != nil {
		return err;
	}

	defer func() {_ = file.Close()}()

	//серилизуем файл - записываем в формат по которому можно будет восстановить структуру
	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err;
	}

	return nil
}

func (s *Storage) PickRandom(userName string) (page *storage.Page, err error) {
	defer func () {err = e.WrapIfErr("can't pick random page", err)} ()

	//получаем путь до директории с файлами
	path := filepath.Join(s.basePath, userName)

	//получаем список файлов
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	//если нету файлов
	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	//получаем случайное число от 0 до номера последнего файла
	n := rand.Intn(len(files))

	file := files[n]

	path = filepath.Join(path, file.Name())
	//декодируем файл и возвращаем его содержимое
	return s.decodePage(path)
}

func (s *Storage) Remove(p *storage.Page) error {
	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap(fmt.Sprintf("can't get file %s for remove", fileName), err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	//удаляем файл
	if err := os.Remove(path); err != nil {
		return e.Wrap(fmt.Sprintf("can't remove file %s", path), err)
	}

	return nil
}

func (s *Storage) IsExists(p *storage.Page) (bool, error) {
	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap(fmt.Sprintf("can't get file %s for remove", fileName), err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	//проверяем существование файла
	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		//файл не существует
		return false, nil
	case err != nil:
		//даже не смошли проверить существование файла
		return false,  e.Wrap(fmt.Sprintf("can't check if exist file %s", fileName), err)
	}

	return true, nil
}

func (s *Storage) decodePage(filePath string) (*storage.Page, error) {
	//открываем файл
	f, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	defer func() {_ = f.Close()} ()

	//переменная в которую декодируем файл
	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

func fileName(p *storage.Page) (string, error){
	return p.Hash()
}