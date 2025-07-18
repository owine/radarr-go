-- name: GetMovieByID :one
SELECT
    id, tmdb_id, imdb_id, title, sort_title,
    year, runtime, overview, path, quality_profile_id,
    monitored, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE id = $1;

-- name: GetAllMovies :many
SELECT
    id, tmdb_id, imdb_id, title, sort_title,
    year, runtime, overview, path, quality_profile_id,
    monitored, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
ORDER BY sort_title;

-- name: GetMoviesByQualityProfile :many
SELECT
    id, tmdb_id, imdb_id, title, sort_title,
    year, runtime, overview, path, quality_profile_id,
    monitored, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE quality_profile_id = $1
ORDER BY sort_title;

-- name: GetMonitoredMovies :many
SELECT
    id, tmdb_id, imdb_id, title, sort_title,
    year, runtime, overview, path, quality_profile_id,
    monitored, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE monitored = true
ORDER BY sort_title;

-- name: CreateMovie :one
INSERT INTO movies (
    tmdb_id, imdb_id, title, sort_title,
    year, runtime, overview, path, quality_profile_id,
    monitored, status, has_file, folder_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING id;

-- name: UpdateMovie :exec
UPDATE movies SET
    tmdb_id = $2,
    imdb_id = $3,
    title = $4,
    sort_title = $5,
    year = $6,
    runtime = $7,
    overview = $8,
    path = $9,
    quality_profile_id = $10,
    monitored = $11,
    status = $12,
    has_file = $13,
    folder_name = $14,
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
    id, tmdb_id, imdb_id, title, sort_title,
    year, runtime, overview, path, quality_profile_id,
    monitored, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE has_file = true
ORDER BY sort_title;
