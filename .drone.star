def main(ctx):
  return {
    "kind": "pipeline",
    "type": "kubernetes",
    "steps": [
      {
        "name": "test",
        "image": "golang",
        "commands": [
            "go test ./... -v -cover"
        ]
      },
      {
        "name": "build",
        "image": "golang",
        "commands": [
            "go build ."
        ]
      }
    ]
  }
