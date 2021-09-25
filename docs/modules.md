TestContainers Java has the concept of
[modules](https://www.testcontainers.org/modules/nginx/).

Those modules are reusable and pre-defined Container for popular projects.

I don't think we are ready to do the same for Go just yet. I proposed a similar
concept called `canned`, mainly to avoid misunderstanding with [Go
modules](https://blog.golang.org/using-go-modules).

* [Canned postgres #98](https://github.com/testcontainers/testcontainers-go/pull/98)
* [Canned minio #105](https://github.com/testcontainers/testcontainers-go/pull/105)
* [Canned MongoDB #102](https://github.com/testcontainers/testcontainers-go/pull/102)
* [Canned Kafka #356](https://github.com/testcontainers/testcontainers-go/pull/356)

Submodules in Go are not easy to maintain today and I want to keep
testcontainers dependencies under control and the project light. If we start
merging modules as part of the main repository we will get one dependency at
least for each of them.  This will blow up quickly adding overheard to the
maintainers of this project as well.

## Share your modules

I know it is important to share code because it makes a project battle tested.
You can create your own modules and you can list them here, in this way other
users will be able to import and reuse your work and experience in the meantime
that we figure out a better solution based as well on how many modules we will
get.
