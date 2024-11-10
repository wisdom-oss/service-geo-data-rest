-- name: get-layers
SELECT *
FROM geodata.layers;

-- name: get-layer
SELECT *
FROM geodata.layers
WHERE
    id = $1;

-- name: get-layer-by-url-key
SELECT *
FROM geodata.layers
WHERE
    "table" = $1;

-- name: get-layer-contents
SELECT id, st_transform(geometry, 4326) AS geometry, key, name, additional_properties
FROM geodata."%s";

-- name: get-layer-object-by-key
SELECT id, st_transform(geometry, 4326) AS geometry, key, name, additional_properties
FROM geodata."%s"
WHERE key = $1::text;

-- name: crate-layer-definition
INSERT INTO geodata.layers(name, description, "table", crs, attribution)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, "table", crs, attribution;

-- name: create-layer-table
CREATE TABLE geodata.%s
(
    id                    bigserial PRIMARY KEY NOT NULL,
    geometry              geometry              NOT NULL,
    key                   text                  NOT NULL,
    name                  text                  NOT NULL,
    additional_properties jsonb
);

-- name: update-geometry-srid
SELECT UpdateGeometrySRID('geodata', $1, 'geometry', $2);