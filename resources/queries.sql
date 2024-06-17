-- name: get-layers
SELECT *
FROM geodata.layers;

-- name: get-layer
SELECT *
FROM geodata.layers
WHERE
    id = $1;

-- name: get-layer-contents
SELECT id, st_transform(geometry, 4326) AS geometry, key, name, additional_properties
FROM geodata.%s;

-- name: crate-layer-definition
INSERT INTO geodata.layers(name, description, "table", crs)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, "table", crs;

-- name: create-layer-table
CREATE TABLE geodata.%s
(
    id                    bigserial PRIMARY KEY NOT NULL,
    geometry              geometry              NOT NULL,
    key                   text                  NOT NULL,
    name                  text                  NOT NULL,
    additional_properties jsonb
);

-- name: insert-shape-object
INSERT INTO geodata.%s(geometry, key, name, additional_properties)
VALUES ($1, $2, $3, $4);