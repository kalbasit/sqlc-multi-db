-- name: GetBook :one
-- SELECT "id", "title", "author", "description", "created_at", "updated_at" FROM books WHERE id = $1
SELECT "id", "title", "author", "description", "created_at", "updated_at"
FROM books
WHERE id = $1;

-- name: ListBooks :many
-- SELECT "id", "title", "author", "description", "created_at", "updated_at" FROM books ORDER BY title
SELECT "id", "title", "author", "description", "created_at", "updated_at"
FROM books
ORDER BY title;

-- name: GetBooksByAuthor :many
-- SELECT "id", "title", "author", "description", "created_at", "updated_at" FROM books WHERE author = $1
SELECT "id", "title", "author", "description", "created_at", "updated_at"
FROM books
WHERE author = $1;

-- name: CreateBook :one
-- INSERT INTO books ("title", "author", "description") VALUES ($1, $2, $3) RETURNING "id", "title", "author", "description", "created_at", "updated_at"
INSERT INTO books ("title", "author", "description")
VALUES ($1, $2, $3)
RETURNING "id", "title", "author", "description", "created_at", "updated_at";

-- name: UpdateBook :one
-- UPDATE books SET "title" = $1, "author" = $2, "description" = $3, "updated_at" = NOW() WHERE id = $4 RETURNING "id", "title", "author", "description", "created_at", "updated_at"
UPDATE books
SET "title" = $1,
    "author" = $2,
    "description" = $3,
    "updated_at" = NOW()
WHERE id = $4
RETURNING "id", "title", "author", "description", "created_at", "updated_at";

-- name: DeleteBook :exec
-- DELETE FROM books WHERE id = $1
DELETE FROM books
WHERE id = $1;

-- name: CreateTag :one
-- INSERT INTO tags ("name") VALUES ($1) RETURNING "id", "name", "created_at", "updated_at"
INSERT INTO tags ("name")
VALUES ($1)
RETURNING "id", "name", "created_at", "updated_at";

-- name: GetTag :one
-- SELECT "id", "name", "created_at", "updated_at" FROM tags WHERE id = $1
SELECT "id", "name", "created_at", "updated_at"
FROM tags
WHERE id = $1;

-- name: AddBookTag :exec
-- INSERT INTO book_tags ("book_id", "tag_id") VALUES ($1, $2)
INSERT INTO book_tags ("book_id", "tag_id")
VALUES ($1, $2);

-- name: AddBookTags :exec
-- INSERT INTO book_tags ("book_id", "tag_id") VALUES (unnest($1::bigint[]), unnest($2::bigint[]))
-- @bulk-for AddBookTag
INSERT INTO book_tags ("book_id", "tag_id")
VALUES (unnest($1::bigint[]), unnest($2::bigint[]));

-- name: GetBookTags :many
-- SELECT t."id", t."name", t."created_at", t."updated_at" FROM tags t INNER JOIN book_tags bt ON t.id = bt.tag_id WHERE bt.book_id = $1
SELECT t."id", t."name", t."created_at", t."updated_at"
FROM tags t
INNER JOIN book_tags bt ON t.id = bt.tag_id
WHERE bt.book_id = $1;
