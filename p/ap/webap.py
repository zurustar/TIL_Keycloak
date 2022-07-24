from flask import Flask
from flask_oidc import OpenIDConnect

app = Flask(__name__)
oidc = OpenIDConnect(app)

@app.get("/")
def index():
    return app.render_template("index.html")

app.run(host="0.0.0.0", port=8080, debug=True)
