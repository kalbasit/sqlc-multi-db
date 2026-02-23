-- name: GetBook :one
-- SELECT "id", "title", "author", "description", "created_at", "updated_at" FROM books WHERE id = ?
SELECT "id", "title", "author", "description", "created_at", "updated_at"
FROM books
WHERE id = ?;

-- name: ListBooks :many
-- SELECT "id", "title", "author", "description", "created_at", "updated_at" FROM books ORDER BY title
SELECT "id", "title", "author", "description", "created_at", "updated_at"
FROM books
ORDER BY title;

-- name: GetBooksByAuthor :many
-- SELECT "id", "title", "author", "description", "created_at", "updated_at" FROM books WHERE author = ?
SELECT "id", "title", "author", "description", "created_at", "updated_at"
FROM books
WHERE author = ?;

-- name: CreateBook :one
-- INSERT INTO books ("title", "author", "description") VALUES (?, ?, ?) RETURNING "id", "title", "author", "description", "created_at", "updated_at"
INSERT INTO books ("title", "author", "description")
VALUES (?, ?, ?)
RETURNING "id", "title", "author", "description", "created_at", "updated_at";

-- name: UpdateBook :one
-- UPDATE books SET "title" = ?, "author" = ?, "description" = ?, "updated_at" = CURRENT_TIMESTAMP WHERE id = ? RETURNING "id", "title", "author", "description", "created_at", "updated_at"
UPDATE books
SET "title" = ?,
    "author" = ?,
    "description" = ?,
    "updated_at" = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING "id", "title", "author", "description", "created_at", "updated_at";

-- name: DeleteBook :exec
-- DELETE FROM books WHERE id = ?
DELETE FROM books
WHERE id = ?;

-- name: CreateTag :one
-- INSERT INTO tags ("name") VALUES (?) RETURNING "id", "name", "created_at", "updated_at"
INSERT INTO tags ("name")
VALUES (?)
RETURNING "id", "name", "created_at", "updated_at";

-- name: GetTag :one
-- SELECT "id", "name", "created_at", "updated_at" FROM tags WHERE id = ?
SELECT "id", "name", "created_at", "updated_at"
FROM tags
WHERE id = ?;

-- name: AddBookTag :exec
-- INSERT INTO book_tags ("book_id", "tag_id") VALUES (?, ?)
INSERT INTO book_tags ("book_id", "tag_id")
VALUES (?, ?);

-- name: AddBookTags :exec
-- INSERT INTO book_tags ("book_id", "tag_id") VALUES (?, ?)
-- @bulk-for AddBookTag
INSERT INTO book_tags ("book_id", "tag_id")
VALUES (?, ?);

-- name: GetBookTags :many
-- SELECT t."id", t."name", t."created_at", t."updated_at" FROM tags t INNER JOIN book_tags bt ON t.id = bt.tag_id WHERE bt.book_id = ?
SELECT t."id", t."name", t."created_at", t."updated_at"
FROM tags t
INNER JOIN book_tags bt ON t.id = bt.tag_id
WHERE bt.book_id = ?;
