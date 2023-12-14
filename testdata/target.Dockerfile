FROM docker.io/alpine AS target0
CMD ["echo", "target0"]

FROM target0 AS target1
CMD ["echo", "target1"]

FROM target1 AS target2
CMD ["echo", "target2"]
