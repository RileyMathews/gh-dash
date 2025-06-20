from dataclasses import dataclass, field
from dataclasses_json import dataclass_json, Undefined, DataClassJsonMixin, config
from datetime import datetime
from marshmallow import fields
from typing import List

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class GithubUser(DataClassJsonMixin):
    login: str
    id: int
    type: str

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class GithubRepo(DataClassJsonMixin):
    id: int
    name: str
    full_name: str
    private: bool
    owner: GithubUser

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class Base(DataClassJsonMixin):
    repo: GithubRepo

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class PrDetail(DataClassJsonMixin):
    title: str
    number: int
    base: Base
    commits_url: str
    comments_url: str
    created_at: datetime = field(
        metadata=config(
            encoder=datetime.isoformat,
            decoder=datetime.fromisoformat,
            mm_field=fields.DateTime(format='iso')
        )
    )

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class Comment(DataClassJsonMixin):
    user: GithubUser
    created_at: datetime = field(
        metadata=config(
            encoder=datetime.isoformat,
            decoder=datetime.fromisoformat,
            mm_field=fields.DateTime(format='iso')
        )
    )

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class CommitAuthor:
    name: str
    email: str
    date: datetime = field(
        metadata=config(
            encoder=datetime.isoformat,
            decoder=datetime.fromisoformat,
            mm_field=fields.DateTime(format='iso')
        )
    )

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class Commit:
    author: CommitAuthor

@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class PrCommit(DataClassJsonMixin):
    author: GithubUser
    commit: Commit

# A top level container that holds the pull request as well as separate nested
# objects for its comments and commits. This is done as each nested item
# is an independant fetch from github and it makes it easier to serialize it
# all straight into their nested classes before getting wrapped up here.
@dataclass_json(undefined=Undefined.EXCLUDE)
@dataclass
class PullRequest(DataClassJsonMixin):
    pr: PrDetail
    comments: List[Comment]
    commits: List[PrCommit]
