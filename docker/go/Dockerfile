# ビルド用の環境
FROM golang:1.18-alpine as builder

# go installにgitを使っているようなのでadd
RUN apk update && apk --no-cache add git && \
    apk --no-cache add vim

# ディレクトリの作成
RUN mkdir /go/app

# 使用しているパッケージのインストール
# RUN go install github.com/bwmarrin/discordgo@latest

# ワーキングディレクトリの設定
WORKDIR /go/app

# ホストのファイルをコンテナの作業ディレクトリに移行
COPY ./app /go/app

# buildを超えてキャッシュを保持し、ファイルが変更されていなければキャッシュの利用ができる
RUN --mount=type=cache,target=~/.cache/go-build \
    CGO_ENABLED=0 \
    go build -installsuffix cgo -o main.build main.go

# 実行ファイルのみの環境
FROM gcr.io/distroless/static as dev

# ビルド用の環境から実行ファイルのみをコピー
COPY --from=builder /go/app/main.build /go/main.build

# 実行
CMD [ "/go/main.build" ]
