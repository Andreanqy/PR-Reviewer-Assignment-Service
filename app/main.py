from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from typing import Union
from db import *
from structures import *

app = FastAPI(title="PR Reviewer Assignment Service")

"""Team endpoints"""
# Создать новую команду
@app.post(
        "/team/add",
        responses = {
            201: {"model": TeamAdd_SuccessfulResponse},
            400: {"model": BadResponse}
        },
        status_code = 201)
async def create_team(request: CreateTeamRequest) -> Union[TeamAdd_SuccessfulResponse, BadResponse]:
    result = team_add(request.team_name, request.members)
    if result["success"]:
        return JSONResponse(status_code=201, content=result["data"])
    else:
        raise HTTPException(status_code=400, detail=result["data"])

# Получить команду с участниками
@app.get(
        "/team/get",
        responses = {
            200: {"model": TeamGet_SuccessfulResponse},
            404: {"model": BadResponse}
        })
async def get_team(team_name: str):
    result = team_get(team_name)
    if result["success"]:
        return JSONResponse(status_code=200, content=result["data"])
    else:
        raise HTTPException(status_code=404, detail=result["data"])

"""User endpoints"""
# Установить флаг активности пользователя
@app.patch(
        "/users/setIsActive",
        responses = {
            200: {"model": User},
            404: {"model": BadResponse}
        })
async def set_user_active(request: SetUserActiveRequest):
    result = user_set_is_active(request.user_id, request.is_active)
    if result["success"]:
        return JSONResponse(status_code=200, content=result["data"])
    else:
        raise HTTPException(status_code=404, detail=result["data"])

# Получить PR'ы, где пользователь назначен ревьювером
@app.get(
        "/users/getReview",
        responses = {
            200: {"model": UserGetReview_SuccessfulResponse}
        })
async def get_user_reviews(user_id: str):
    result = user_get_review(user_id)
    return JSONResponse(status_code=200, content=result["data"])

"""PR endpoints"""
# Создать PR и автоматически назначить до 2 ревьюверов из команды автора
@app.post(
        "/pullRequest/create",
        responses = {
            201: {"model": PRCreate_SuccessfulResponse},
            404: {"model": BadResponse},
            409: {"model": BadResponse}
        },
        status_code = 201)
async def create_pull_request(request: PRShort):
    result = pr_create(request.pull_request_id, request.pull_request_name, request.author_id)
    if result["success"]:
        return JSONResponse(status_code=201, content=result["data"])
    else:
        if result["data"]["code"] == "NOT_FOUND":
            raise HTTPException(status_code=404, detail=result["data"])
        elif result["data"]["code"] == "PR_EXISTS":
            raise HTTPException(status_code=409, detail=result["data"])

# Пометить PR как MERGED
@app.post(
        "/pullRequest/merge",
        responses = {
            200: {"model": PRMerge_SuccessfulResponse},
            404: {"model": BadResponse}
        })
async def merge_pull_request(pull_request_id: str):
    result = pr_merge(pull_request_id)
    if result["success"]:
        return JSONResponse(status_code=200, content=result["data"])
    else:
        raise HTTPException(status_code=404, detail=result["data"])

# Переназначить конкретного ревьювера на другого из его команды
@app.post(
        "/pullRequest/reassign",
        responses = {
            200: {"model": PRReassign_SuccessfulResponse},
            404: {"model": BadResponse}
        })
async def reassign_reviewer(request: ReassignPRRequest):
    result = pr_reassign(request.pull_request_id, request.old_user_id)
    if result["success"]:
        return JSONResponse(status_code=200, content=result["data"])
    else:
        raise HTTPException(status_code=404, detail=result["data"])

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)