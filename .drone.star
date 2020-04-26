def main(ctx):
  return {
    "kind": "pipeline",
    "type": "kubernetes",
    "steps": [
      {
        "name": "test & build",
        "image": "l.gcr.io/google/bazel:latest",
        "commands": [
          "echo 'build --disk_cache=/cache/funcheck' >> .bazelrc",
          "bazel version",
          "bazel test --color=yes ...",
          "bazel build --color=yes ..."
        ]
      }
    ]
  }
