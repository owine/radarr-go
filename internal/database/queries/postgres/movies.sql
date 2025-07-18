-- name: GetMovieByID :one
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE id = $1;

-- name: GetAllMovies :many
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
ORDER BY sort_title;

-- name: GetMoviesByQualityProfile :many
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE quality_profile_id = $1
ORDER BY sort_title;

-- name: GetMonitoredMovies :many
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE monitored = true
ORDER BY sort_title;

-- name: CreateMovie :one
INSERT INTO movies (
    tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    folder_name, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    NOW(), NOW()
) RETURNING id;

-- name: UpdateMovie :exec
UPDATE movies SET
    tmdb_id = $2,
    imdb_id = $3,
    title = $4,
    original_title = $5,
    sort_title = $6,
    year = $7,
    release_date = $8,
    runtime = $9,
    certification = $10,
    overview = $11,
    images = $12,
    genres = $13,
    studio = $14,
    path = $15,
    quality_profile_id = $16,
    monitored = $17,
    minimum_availability = $18,
    status = $19,
    has_file = $20,
    added = $21,
    folder_name = $22,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteMovie :exec
DELETE FROM movies WHERE id = $1;

-- name: CountMovies :one
SELECT COUNT(*) FROM movies;

-- name: CountMonitoredMovies :one
SELECT COUNT(*) FROM movies WHERE monitored = true;

-- name: GetMoviesWithFiles :many
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE has_file = true
ORDER BY sort_title;
