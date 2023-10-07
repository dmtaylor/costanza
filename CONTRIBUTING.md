# Contributing Guide

Thanks for being interested in contributing to this project. The purpose of this doc is to set out guidelines to
smooth out the contribution process & set expectations in terms of contributions. This is a hobby project of mine, and
I don't expect much in the way of contributions. If you are interested in making changes, please make a GitHub issue
describing what you're trying to do & why, and I would be happy to review any pull requests. Since this is a hobby
project, I don't have any guarantees on turnaround for reviews, and I reserve the right to reject any changes. I also
reserve the right to make changes unilaterally, without making a GitHub issue or communicating to other parties.

## Guidelines
* Don't be a jerk. Go into making changes with love, not hatred.
* Run `go fmt` on all go files that are changed before committing.
* Group imports into three ordered groups: standard libraries, external dependencies, then internal dependencies.
E.g.
 
        package example
        import (
            "log/slog"
            "os"

            "github.com/prometheus/client_golang/prometheus"

            "github.com/dmtaylor/costanza/internal/util"
        )
        ...
* Verify that all tests pass before opening a review.
* Tend towards writing tests for any added code. There are definitely cases where adding tests is impractical: a general
guideline is "if the code doesn't make a networked API call, it should have a test." The target minimum test coverage
is 75%.