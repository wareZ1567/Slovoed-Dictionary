package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	// go get github.com/mattn/go-sqlite3 для скачивания драйвера SQLite
	_ "github.com/mattn/go-sqlite3"
)

type Wm struct {
	Word    string
	Meaning string
}

var dbp *sql.DB // Для вывода словаря на веб-страницу

func main() {
	dbName := "mainDB.db"

	createIfNotExist(dbName)

	db, err := sql.Open("sqlite3", dbName) // Открыть БД
	if err != nil {
		fmt.Println("ОШИБКА открытия БД.")
	}
	dbp = db

	defer closeDB(db) // Закрыть БД

	// Создание таблицы, если её нет
	sqlCommand("CREATE TABLE IF NOT EXISTS dictionary (word TEXT PRIMARY KEY NOT NULL, meaning TEXT NOT NULL);", db)

	// Вывод словаря в консоль
	printDict(db)

	var option string

	// Текстовое меню с выбором действий со словарём
	for {
		fmt.Println("Введите N , чтобы ввести новое слово.\nQ для выхода.\n" +
			"P для вывода словаря в консоль в алфавитном порядке.\nS для сохранения на диск.\n" +
			"W Для вывода словаря на веб-страницу.\nD для удаления словаря.\n")
		_, err := fmt.Scan(&option)
		if err != nil {
			fmt.Println("какая-то ошибка в меню")
		}

		switch option {
		case "Q":
			{
				break
			}
		case "N":
			{
				insertWordAndMeaning(db)
			}
		case "P":
			{
				printDict(db)
			}
		case "S":
			{
				saveToFile(db)
			}
		case "W":
			{
				server := http.Server{
					Addr:    ":5555",
					Handler: http.HandlerFunc(router),
				}

				if err := server.ListenAndServe(); err != nil {
					fmt.Println("Ошибка запуска сервера")
				}
			}
		case "D":
			{
				var answ string
				for {
					fmt.Println("Вы уверены, что хотите удалить словарь?\nY = да, N = не удалять.")
					_, err = fmt.Scan(&answ)
					if err != nil {
						fmt.Println(err)
					}
					switch answ {
					case "Y":
						{
							sqlCommand("DELETE FROM dictionary;", db)
							break
						}
					case "N":
						{
							break
						}
					}
					if answ == "N" || answ == "Y" {
						break
					}
				}
			}
		}
		if option == "Q" {
			break
		}
	}
}

// Маршрутизация обработчиков по запрашиваемым адресам
func router(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		printDictToWebPage(w, r, dbp)
	default:
		pageNotFound404(w, r)
	}
}

// Вывод словаря на веб-страницу
func printDictToWebPage(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	res, err := db.Query("SELECT * FROM dictionary ORDER BY word")
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err = res.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var pair Wm
	var s string
	for res.Next() {
		err = res.Scan(&pair.Word, &pair.Meaning)
		s = fmt.Sprintf("%v = %v\n", pair.Word, pair.Meaning)
		_, err = fmt.Fprintf(w, s)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func pageNotFound404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound) // 404
	return
}

// Сохраняет словарь в файл на локальном хранилище, имя файла задаёт пользователь
func saveToFile(db *sql.DB) {
	var textFileName string

	fmt.Println("Введите имя файла для записи словаря. Расширением файла будет \".TXT\" .")
	_, err := fmt.Scan(&textFileName)
	if err != nil {
		fmt.Println(err)
	}
	textFileName += ".txt"

	// Создать файл
	file, err := os.Create(textFileName)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// Получение всех значений записей в таблице БД, в которой хранится словарь
	res, err := db.Query("SELECT * FROM dictionary ORDER BY word")
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err = res.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var pair Wm
	var s string
	// Запись словаря в файл построчно каждой пары "слово - значение"
	for res.Next() {
		err = res.Scan(&pair.Word, &pair.Meaning)
		s = fmt.Sprintf("%v = %v\n", pair.Word, pair.Meaning)
		_, err := file.WriteString(s)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// Добавление в таблицу БД нового слова и значения
func insertWordAndMeaning(db *sql.DB) {
	var pair Wm
	fmt.Println("Введите слово и нажмите ENTER.")
	_, err := fmt.Scan(&pair.Word)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Введите значение и нажмите ENTER.")
	_, err = fmt.Scan(&pair.Meaning)
	if err != nil {
		fmt.Println(err)
	}

	var sqlString string
	sqlString = fmt.Sprintf("INSERT INTO dictionary (word, meaning) VALUES ('%v', '%v');", pair.Word, pair.Meaning)
	sqlCommand(sqlString, db)
}

// Создаёт файл, если его не существует
func createIfNotExist(fileName string) {
	file, err := os.OpenFile(fileName, os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
	}

	err = file.Close()
	if err != nil {
		fmt.Println(err)
	}
}

// Распечатывает словарь в консоль в алфавитном порядке
func printDict(db *sql.DB) {
	res, err := db.Query("SELECT * FROM dictionary ORDER BY word")
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err = res.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var pair Wm
	// Построчно выводит пары "слово - значение" из словаря в консоль
	for res.Next() {
		err = res.Scan(&pair.Word, &pair.Meaning)
		fmt.Printf("%v = %v\n", pair.Word, pair.Meaning)
	}
	err = res.Close()
	if err != nil {
		fmt.Println(err)
	}
}

// Выполняет SQL-инструкции к БД
func sqlCommand(sqlString string, db *sql.DB) {
	stmt, err := db.Prepare(sqlString)
	if err != nil {
		fmt.Println(err)
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Printf("ОШИБКА выполнения SQL команды ф-ей Exec(). %v", err)
	}
	err = stmt.Close()
	if err != nil {
		fmt.Println(err)
	}
}

// Закрывает БД
func closeDB(db *sql.DB) {
	err := db.Close()
	if err != nil {
		fmt.Println(err)
	}
}
