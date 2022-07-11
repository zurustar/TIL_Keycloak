
#
# pip install Flask
# pip install uvicorn
# uvicorn app:app --reload
#

from flask import Flask
app = Flask("dummy")

@app.get('/')
def hello():
    return {"message": "Ok"}


