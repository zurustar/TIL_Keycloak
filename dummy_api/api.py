
#
# pip install fastapi
# pip install uvicvorn
# uvicorn api:main --port 5000 --reload
#

db_filename = "./dummy.db"

from fastapi import FastAPI
import sqlite3

#
# DB初期化
#
def init_db(filename):
    db = sqlite3.connect(filename)
    cur = db.cursor()
    cur.execute("DROP TABLE IF EXISTS TagTypes;")
    cur.execute("DROP TABLE IF EXISTS Tags;")
    cur.execute("""
        CREATE TABLE TagTypes (
            tag_type_id INTEGER
            , name TEXT
        )""")
    for data in [(1, "会社"), (2, "製品")]:
        cur.execute("""
            INSERT INTO TagTypes(tag_type_id, name) VALUES(?, ?)
            """, data)
    cur.execute("""
        CREATE TABLE Tags (
            tag_id INTEGER
            , tag_type_id INTEGER
            , name TEXT
            , FOREIGN KEY (tag_type_id) REFERENCES TagTypes(tag_type_id)
        )""")
    db.commit()
    db.close()


app = FastAPI()
init_db(db_filename)

@app.get("/")
def get_root():
    return {"Hello": "World"}


@app.get("/tags")
def get_tags():
    db = sqlite3.connect(db_filename)
    data = []
    for row in db.execute("SELECT tag_id, tag_type_id, name FROM Tags"):
        data.append(row)
    db.close()
    return {"tags": data}

@app.get("/tags/{tag_id}")
def get_item(item_id: int, q: str = None):
    db = sqlite3.connect(db_filename)
    for row in db.execute("SELECT tag_id, tag_type_id, name FROM Tags"):
        data.append(row)
    db.close()
    return {"tags": data}

