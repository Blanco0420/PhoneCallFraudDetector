
# === Update and install dependencies ===
FROM gocv/opencv:latest AS base

RUN apt-get update 

RUN apt-get install -y \
  software-properties-common \
  python3 \
  python3-pip \
  git



# === Install libpostal ===
WORKDIR /tmp


RUN git clone https://github.com/openvenues/libpostal libpostal-building && \
  cd libpostal-building && \
  ./bootstrap.sh && \
  mkdir /data && \
  ./configure  && \
  make && \
  make install && \
  ldconfig


# === Install Tesseract ===

RUN apt-get install -y \
  tesseract-ocr \
  libtesseract-dev \
  tesseract-ocr-eng

# === Install Firefox driver ===

RUN wget https://github.com/mozilla/geckodriver/releases/download/v0.36.0/geckodriver-v0.36.0-linux64.tar.gz && \
  tar -xzvf geckodriver* && \
  chmod +x geckodriver && \
  mv geckodriver /usr/local/bin

RUN add-apt-repository ppa:mozillateam/ppa -y && \
  apt-get install -y firefox-esr 

# === Cleanup apt cache ===

RUN rm -rf /var/lib/apt/lists/*

WORKDIR /app


COPY ./pythonBrowsing ./pythonBrowsing

RUN python3 --version

RUN python3 -m pip install -r /app/pythonBrowsing/requirements.txt
RUN python3 -m pip install fastapi[standard]

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download


EXPOSE 8080

COPY ./entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT [ "/app/entrypoint.sh" ]

# === === DEV === ===
FROM base AS dev
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/air-verse/air@latest

EXPOSE 40000

CMD [ "air", "-c", ".air.toml" ]

# === === PROD === ===
FROM base AS prod

COPY . .
RUN go build -o backend
RUN chmod +x /app/backend

CMD [ "/app/backend" ]
