import geoalchemy2
import sqlalchemy

import database

metadata = sqlalchemy.MetaData(schema="geodata")

shapes = sqlalchemy.Table(
    "shapes",
    metadata,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True, autoincrement=True),
    sqlalchemy.Column("name", sqlalchemy.Text),
    sqlalchemy.Column("key", sqlalchemy.Text),
    sqlalchemy.Column("nuts_key", sqlalchemy.Text),
    sqlalchemy.Column("geom", geoalchemy2.Geometry("MultiPolygon")),
)
