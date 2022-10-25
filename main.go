package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var schema = `
CREATE TABLE IF NOT EXISTS article (
	id CHAR(36) PRIMARY KEY,
	title VARCHAR(255) UNIQUE NOT NULL,
	body TEXT NOT NULL,
	author_id CHAR(36),
	created_at TIMESTAMP DEFAULT now(),
	updated_at TIMESTAMP,
	deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS author (
	id CHAR(36) PRIMARY KEY,
	firstname VARCHAR(255) NOT NULL,
	lastname VARCHAR(255) NOT NULL,
	created_at TIMESTAMP DEFAULT now(),
	updated_at TIMESTAMP,
	deleted_at TIMESTAMP
);

ALTER TABLE article DROP CONSTRAINT IF EXISTS fk_article_author;
ALTER TABLE article ADD CONSTRAINT fk_article_author FOREIGN KEY (author_id) REFERENCES author (id);
`

// Article ...
type Article struct {
	ID        string     `db:"id"`
	Title     string     `db:"title"`
	Body      string     `db:"body"`
	AuthorID  string     `db:"author_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// Author ...
type Author struct {
	ID        string     `db:"id"`
	Firstname string     `db:"firstname"`
	Lastname  string     `db:"lastname"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func main() {
	db, err := sqlx.Connect("postgres", "user=admin password=qwerty123 dbname=article_db sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(schema)

	tx := db.MustBegin()

	tx.MustExec("INSERT INTO author (id, firstname, lastname) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", "3e1dfc06-dcf6-41fc-b3cc-7c0563fdfab3", "John", "Doe")
	tx.MustExec("INSERT INTO author (id, firstname, lastname) VALUES ($3, $2, $1) ON CONFLICT DO NOTHING", "Botirov", "Saidamir", "24000e82-9c48-4297-a442-ecd1ad55791e")

	tx.MustExec("INSERT INTO article (id, title, body, author_id) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", "26e2aebc-9771-45ba-8577-ef1a2e7b4170", "Lorem 1", "Body 1", "3e1dfc06-dcf6-41fc-b3cc-7c0563fdfab3")
	tx.MustExec("INSERT INTO article (id, title, body, author_id) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", "9900756f-e3ed-4dd7-a3a8-4e3cef248ccc", "Lorem 2", "Body 2", "24000e82-9c48-4297-a442-ecd1ad55791e")

	tx.NamedExec("INSERT INTO article (id, title, body, author_id) VALUES (:id, :t, :b, :aid) ON CONFLICT DO NOTHING", map[string]interface{}{
		"id":  "3e451dc4-42e8-4dbc-a70b-edee8f6452ba",
		"t":   "Lorem 3",
		"b":   "Body 3",
		"aid": "3e1dfc06-dcf6-41fc-b3cc-7c0563fdfab3",
	})

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	var authorList []Author
	err = db.Select(&authorList, "SELECT * FROM author ORDER BY created_at ASC")
	if err != nil {
		panic(err)
	}

	fmt.Printf("john -> %#v\n", authorList[0])
	fmt.Printf("me -> %#v\n", authorList[1])

	var article1 Article
	err = db.Get(&article1, "SELECT * FROM article WHERE id=$1", "26e2aebc-9771-45ba-8577-ef1a2e7b4170")
	fmt.Printf("article1 -> %#v\n", article1)

	rows, err := db.Queryx("SELECT id, title, body, author_id, created_at, updated_at, deleted_at FROM article")
	if err != nil {
		panic(err)
	}

	fmt.Printf("rows -> %#v\n", rows)

	var articleList []Article
	i := 0
	for rows.Next() {
		var a Article

		err := rows.Scan(
			&a.ID,
			&a.Title,
			&a.Body,
			&a.AuthorID,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		)
		// err := rows.StructScan(&a)
		if err != nil {
			panic(err)
		}
		i++
		fmt.Printf("%d a->%#v\n", i, a)
		articleList = append(articleList, a)
	}

	fmt.Printf("articleList -> %#v\n", articleList)

	res, err := db.NamedExec("UPDATE article  SET title=:t, body=:b, updated_at=now() WHERE id=:id", map[string]interface{}{
		"id": "3e451dc4-42e8-4dbc-a70b-edee8f6452ba",
		"t":  "Lorem updated",
		"b":  "Body updated",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("res -> %#v\n", res)

	n, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	fmt.Printf("res.RowsAffected() -> %#v\n", n)

	var a Article
	err = db.QueryRow(`SELECT id, title, body, author_id, created_at, updated_at, deleted_at FROM article WHERE id = $1`, "3e451dc4-42e8-4dbc-a70b-edee8f6452ba").Scan(
		&a.ID,
		&a.Title,
		&a.Body,
		&a.AuthorID,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.DeletedAt,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("a -> %#v\n", a)

	res, err = db.Exec("DELETE FROM article WHERE id=$1", "9900756f-e3ed-4dd7-a3a8-4e3cef248ccc")
	if err != nil {
		panic(err)
	}

	n, err = res.RowsAffected()
	if err != nil {
		panic(err)
	}
	fmt.Printf("res.RowsAffected() -> %#v\n", n)
}
