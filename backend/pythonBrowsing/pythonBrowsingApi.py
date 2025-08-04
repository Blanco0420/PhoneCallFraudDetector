import cloudscraper
from fastapi import FastAPI, Response
from fastapi.responses import HTMLResponse

app = FastAPI()

scraper = cloudscraper.create_scraper()

@app.get("/getNumberPage")
def getNumberPage(url: str):
    content = scraper.get(url)
    if content.ok:
        return HTMLResponse(content=content.text)
    else:
        return Response(status_code=500)
