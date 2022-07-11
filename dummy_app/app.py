from flask import Flask
app = Flask("dummy")

@app.get('/')
def hello():
    return {"message": "Ok"}

if __name__ == '__main__':
    app.run(debug=True)
