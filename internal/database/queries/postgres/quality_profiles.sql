-- name: GetQualityProfileByID :one
SELECT id, name, cutoff, items, created_at, updated_at
FROM quality_profiles
WHERE id = $1;

-- name: GetAllQualityProfiles :many
SELECT id, name, cutoff, items, created_at, updated_at
FROM quality_profiles
ORDER BY name;

-- name: CreateQualityProfile :one
INSERT INTO quality_profiles (name, cutoff, items, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id;

-- name: UpdateQualityProfile :exec
UPDATE quality_profiles SET
    name = $2,
    cutoff = $3,
    items = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteQualityProfile :exec
DELETE FROM quality_profiles WHERE id = $1;

-- name: CountQualityProfiles :one
SELECT COUNT(*) FROM quality_profiles;
