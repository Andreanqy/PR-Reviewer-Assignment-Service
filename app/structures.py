"""
from typing import List, Dict
from enum import Enum

class ErrorResponse(TypedDict):
    code: code_error
    message: str

class TeamMember(TypedDict):
    user_id: str
    username: str
    is_active: bool

class Team(TypedDict):
    team_name: str
    members: List [TeamMember]

class User(TypedDict):
    user_id: str
    username: str
    team_name: str
    is_active: bool

class PullRequest(TypedDict):
    pull_request_id: str
    pull_request_name: str
    author_id: str
    status: str
    assigned_reviewers: List[str]
    createdAt: str
    mergedAt: str

class PullRequestShort(TypedDict):
    pull_request_id: str
    pull_request_name: str
    author_id: str
    status: str
"""



from enum import Enum
from typing import List
from pydantic import BaseModel



class code_error(str, Enum):
    TEAM_EXISTS = "TEAM_EXISTS"
    PR_EXISTS = "PR_EXISTS"
    PR_MERGED = "PR_MERGED"
    NOT_ASSIGNED = "NOT_ASSIGNED"
    NO_CANDIDATE = "NO_CANDIDATE"
    NOT_FOUND = "NOT_FOUND"

class TeamMember(BaseModel):
    user_id: str
    username: str
    is_active: bool = True

class User(BaseModel):
    user_id: str
    username: str
    team_name: str
    is_active: bool = True

class PRShort(BaseModel):
    pull_request_id: str
    pull_request_name: str
    author_id: str
    status: str



"""=== Структуры тела запросов"""

class CreateTeamRequest(BaseModel):
    team_name: str
    members: List[TeamMember]

#class GetTeamRequest(BaseModel):
#    team_name: str

class SetUserActiveRequest(BaseModel):
    user_id: str
    is_active: bool

#class GetUserReviewsRequest(BaseModel):
#    user_id: str

class CreatePRRequest(PRShort):
    assigned_revieweres: List[str]

#class MergePRRequest(BaseModel):
#    pull_request_id: str

class ReassignPRRequest(BaseModel):
    pull_request_id: str
    old_user_id: str



"""=== Структуры тела ответов ==="""

class BadResponse(BaseModel):
    code: str
    message: str

class TeamAdd_SuccessfulResponse(BaseModel):
    team_name: str
    members: List[TeamMember]

class TeamGet_SuccessfulResponse(BaseModel):
    team_name: str
    members: List[User]

# Проверить

class UserGetReview_SuccessfulResponse(BaseModel):
    user_id: str
    pull_requests: List[PRShort]

class PRCreate_SuccessfulResponse(PRShort):
    assigned_reviewers: List[str]

class PRMerge_SuccessfulResponse(PRCreate_SuccessfulResponse):
    mergedAt: str

class PRReassign_SuccessfulResponse(BaseModel):
    pr: PRCreate_SuccessfulResponse
    replaced_by: str