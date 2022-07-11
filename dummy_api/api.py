
#
# pip install fastapi
# pip install uvicvorn
# uvicorn api:main --reload
#

from fastapi import FastAPI
app = FastAPI()

@app.get("/")
def get_root():
    return {"Hello": "World"}


@app.get("/items/{item_id}")
def get_item(item_id: int, q: str = None):
    return {"item_id": item_id, "q": q}