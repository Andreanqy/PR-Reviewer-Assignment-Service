import psycopg2
from contextlib import contextmanager
from typing import List, Dict, Any
from structures import TeamMember
from datetime import datetime
import os

def get_connection():
    return psycopg2.connect(
        dbname = os.getenv("POSTGRES_DB", "mydb"),
        user = os.getenv("POSTGRES_USER", "postgres"),
        password = os.getenv("POSTGRES_PASSWORD", "password"),
        host = os.getenv("POSTGRES_HOST", "db"),
        port = os.getenv("POSTGRES_PORT", "5432")
    )

@contextmanager
def get_cursor():
    # conn = psycopg2.connect(dbname="mydb", user="postgres", password="password")
    conn = get_connection()
    cursor = conn.cursor()

    try:
        yield cursor
        conn.commit()
    except:
        conn.rollback()
        raise
    finally:
        cursor.close()
        conn.close()

def update_db_before_test():
    conn = psycopg2.connect(dbname="mydb", user="postgres", password="password")
    cursor = conn.cursor()

    # Удалить pr-1001
    cursor.execute("""
                   DELETE FROM pull_requests
                   WHERE pull_request_id = 'pr-1001';
                   """)
    
    # Удалить pr-1002
    cursor.execute("""
                   DELETE FROM pull_requests
                   WHERE pull_request_id = 'pr-1002';
                   """)
    
    # Удалить u6, u7
    # Удаляю пользователей example_user_6, example_user_7
    cursor.execute("""
                   DELETE FROM users
                   WHERE user_id = ANY(%s);
                   """, (['u6', 'u7'],))
    # Удаляю команду ExampleTeam
    cursor.execute("""
                   DELETE FROM teams
                   WHERE team_name = %s;
                   """, ("ExampleTeam",))
    
    # Установить is_active=True для u1
    cursor.execute("""
                       UPDATE users
                       SET is_active = %s
                       WHERE user_id = %s
                       """, (True, 'u1'))
    
    conn.commit()

def print_db_users():
    with get_cursor() as cursor:
        cursor.execute("""SELECT * FROM users;""")
        print(cursor.fetchall())

def team_add(team_name: str, members: List[TeamMember]) -> Dict[str, Any]:
    with get_cursor() as cursor:
        # Проверка на наличие команды с таким же названием
        cursor.execute("""
                       SELECT EXISTS(
                        SELECT 1 FROM teams WHERE team_name = %s
                       );
                       """, (team_name,))
        if cursor.fetchone()[0]:
            return {
                "success": False,
                "data": {
                    "code": "TEAM_EXISTS",
                    "message": "team_name already exists"
                }
            }
        
        # Создать команду
        cursor.execute("""
                       INSERT INTO teams (team_name, members)
                       VALUES (%s, %s)
                       """, (team_name, [user.user_id for user in members]))

        # Добавить ревьюверов в команду
        for user in members:
            cursor.execute("""
                           INSERT INTO users (user_id, username, team_name, is_active)
                           VALUES (%s, %s, %s, %s);
                           """, (user.user_id, user.username, team_name, user.is_active))
        return {
            "success": True,
            "data": {
                "team": team_name,
                "members": [member.model_dump() for member in members]
            }
        }

def team_get(team_name: str) -> Dict[str, Any]:
    with get_cursor() as cursor:
        # Проверка наличия команды с таким названием
        cursor.execute("""
                        SELECT EXISTS (
                            SELECT 1 FROM teams WHERE team_name = %s
                        );
                        """, (team_name,))
        if not cursor.fetchone()[0]:
            return {
                "success": False,
                "data": {
                    "code": "NOT_FOUND",
                    "message": "team_name not found"
                }
            }

        # Получить ID участников команды
        cursor.execute("""
                       SELECT members FROM teams
                       WHERE team_name = %s
                       """, (team_name,))
        ids = cursor.fetchone()[0]

        users: List[dict] = []
        for id in ids:
            cursor.execute("""
                           SELECT * FROM users
                           WHERE user_id = %s
                           """, (id,))
            user = cursor.fetchone()
            users.append({
                "user_id": user[0],
                "username": user[1],
                "team_name": user[2],
                "is_active": user[3]
            })
        return {
            "success": True,
            "data": {
                "team_name": team_name,
                "members": users
            }
        }

def user_set_is_active(user_id: str, is_active: bool) -> Dict[str, Any]:
    with get_cursor() as cursor:
        # Проверка наличия пользователя с таким ID
        cursor.execute("""
                       SELECT EXISTS (
                        SELECT 1 FROM users WHERE user_id = %s
                       );
                       """, (user_id,))
        if not cursor.fetchone()[0]:
            return {
                "success": False,
                "data": {
                    "code": "NOT_FOUND",
                    "message": "user_id not found"
                }
            }

        cursor.execute("""
                       UPDATE users
                       SET is_active = %s
                       WHERE user_id = %s
                       """, (is_active, user_id))
        cursor.execute("""
                       SELECT * FROM users
                       WHERE user_id = %s
                       """, (user_id,))
        user = cursor.fetchone()
        return {
            "success": True,
            "data": {
                "user_id": user[0],
                "username": user[1],
                "team_name": user[2],
                "is_active": user[3]
            }
        }

def pr_create(pull_request_id: str, pull_request_name: str, author_id: str) -> Dict[str, Any]:
    with get_cursor() as cursor:
        # Проверка на наличие PR с таким же ID
        cursor.execute("""
                       SELECT EXISTS(
                        SELECT 1 FROM pull_requests WHERE pull_request_id = %s
                       );
                       """, (pull_request_id,))
        if cursor.fetchone()[0]:
            return {
                "success": False,
                "data": {
                    "code": "PR_EXISTS",
                    "message": "PR id already exists"
                }
            }
        
        # Проверка на наличие author_id
        cursor.execute("""
                       SELECT EXISTS(
                        SELECT 1 FROM users WHERE user_id = %s
                       );
                       """, (author_id,))
        if not cursor.fetchone()[0]:
            return {
                "success": False,
                "data": {
                    "code": "NOT_FOUND",
                    "message": "user_id not found"
                }
            }

        assigned_reviewers = []
        cursor.execute("""
                       SELECT team_name FROM users
                       WHERE user_id = %s;
                       """, (author_id,))
        author_team = cursor.fetchone()[0]
        cursor.execute("""
                       SELECT user_id FROM users
                       WHERE user_id != %s AND team_name = %s AND is_active = true LIMIT 2;
                       """, (author_id, author_team))
        rows = cursor.fetchall()
        assigned_reviewers = [row[0] for row in rows]
        cursor.execute("""
                       INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, assigned_reviewers, createdAt)
                       VALUES (%s, %s, %s, %s, %s, %s);
                       """, (pull_request_id, pull_request_name, author_id, "OPEN", assigned_reviewers, datetime.now().isoformat()+"Z"))
        return {
            "success": True,
            "data": {
                "pull_request_id":pull_request_id,
                "pull_request_name":pull_request_name,
                "author_id":author_id,
                "status":"OPEN",
                "assigned_reviewers":assigned_reviewers
               }
        }

def pr_merge(pull_request_id: str) -> Dict[str, Any]:
    with get_cursor() as cursor:
        # Проверка на наличие PR
        cursor.execute("""
                       SELECT EXISTS(
                        SELECT 1 FROM pull_requests WHERE pull_request_id = %s
                       );
                       """, (pull_request_id,))
        if not cursor.fetchone()[0]:
            return {
                "success": False,
                "data": {
                    "code": "NOT_FOUND",
                    "message": "pull_request_id not found"
                }
            }
        
        cursor.execute("""
                       UPDATE pull_requests
                       SET
                        status = 'MERGED',
                        mergedAt = %s
                       WHERE pull_request_id = %s
                       RETURNING mergedAt;
                       """, (datetime.now().isoformat()+"Z", pull_request_id))

        cursor.execute("""
                       SELECT * FROM pull_requests
                       WHERE pull_request_id = %s;
                       """, (pull_request_id,))
        pr = cursor.fetchone()
        
        return {
            "success": True,
            "data": {
                "pull_request_id": pr[0],
                "pull_request_name": pr[1],
                "author_id": pr[2],
                "status": pr[3],
                "assigned_reviewers": pr[4],
                "mergedAt": pr[6].isoformat() + "Z"
            }
        }

def pr_reassign(pull_request_id: str, old_user_id: str) -> Dict[str, Any]:
    with get_cursor() as cursor:
        # Проверка на наличие PR
        cursor.execute("""
                        SELECT EXISTS(
                            SELECT 1 FROM pull_requests WHERE pull_request_id = %s
                       );
                       """, (pull_request_id,))
        if not cursor.fetchone()[0]:
            return {
                "success": False,
                "data": {
                    "code": "NOT_FOUND",
                    "message": "pull_request_id not found"
                }
            }
        
        # Проверка на наличие user_id
        cursor.execute("""
                        SELECT EXISTS(
                            SELECT 1 FROM users WHERE user_id = %s
                       );
                       """, (old_user_id,))
        if not cursor.fetchone()[0]:
            return {
                "success": False,
                "data": {
                    "code": "NOT_FOUND",
                    "message": "user_id not found"
                }
            }
        
        # Получить ID автора PR
        cursor.execute("""
                       SELECT author_id FROM pull_requests
                       WHERE pull_request_id = %s;
                       """, (pull_request_id,))
        author_id = cursor.fetchone()[0]

        # Получить название команды PR
        cursor.execute("""
                       SELECT team_name FROM users
                       WHERE user_id = %s;
                       """, (old_user_id,))
        team_name = cursor.fetchone()[0]

        # Получить список ID ревьюверов PR
        cursor.execute("""
                       SELECT assigned_reviewers FROM pull_requests
                       WHERE pull_request_id = %s;
                       """, (pull_request_id,))
        assigned_reviewers = cursor.fetchone()[0]
        not_reviewers = assigned_reviewers.copy()
        not_reviewers.append(author_id)

        # Получить ID нового ревьювера
        cursor.execute("""
                       SELECT user_id FROM users
                       WHERE team_name = %s
                       AND user_id != ALL(%s)
                       AND is_active = true
                       LIMIT 1;
                       """, (team_name, not_reviewers))
        new_user_id = cursor.fetchone()[0]

        assigned_reviewers.remove(old_user_id)
        assigned_reviewers.append(new_user_id)

        # Получить название PR
        cursor.execute("""
                       SELECT pull_request_name FROM pull_requests
                       WHERE pull_request_id = %s;
                       """, (pull_request_id,))
        pull_request_name = cursor.fetchone()[0]

        # Установить новых ревьюверов
        cursor.execute("""
                      UPDATE pull_requests
                      SET assigned_reviewers = %s
                      WHERE pull_request_id = %s;
                      """, (assigned_reviewers, pull_request_id))
        return {
            "success": True,
            "data": {
                "pr": {
                    "pull_request_id":pull_request_id,
                    "pull_request_name":pull_request_name,
                    "author_id":author_id,
                    "status":"OPEN",
                    "assigned_reviewers":assigned_reviewers
                },
                "replaced_by": new_user_id
            }
        }

def user_get_review(user_id: str) -> Dict[str, Any]:
    with get_cursor() as cursor:
        cursor.execute("""
                       SELECT * FROM pull_requests
                       WHERE %s = ANY(assigned_reviewers);
                       """, (user_id,))
        prs = cursor.fetchall()
        answer = []
        for pr in prs:
            answer.append({
                "pull_request_id": pr[0],
                "pull_request_name": pr[1],
                "author_id": pr[2],
                "status": pr[3]
            })
        return {
            "success": True,
            "data": {
                "user_id": user_id,
                "pull_requests": answer
            }
        }

if __name__ == "__main__":
    update_db_before_test()
    
    print("\n")
    print("=== База данных пользователей до работы программы")
    print_db_users()
    print("\n")

    print("=== Добавляем PR №1 ===")
    print(pr_create("pr-1001", "Add new feature", "u1"))
    print("\n")
    
    print("=== Добавляем PR №2 ===")
    print(pr_create("pr-1002", "Fix some bugs", "u2"))
    print("\n")

    print("=== Заменяем ревьювера u2 у PR pr-1001 ===")
    print(pr_reassign('pr-1001', 'u2'))
    print("\n")

    """
    print("=== Добавляем в команду ExampleTeam новых пользователей ===")
    members: List[Dict[str, Any]] = [
        {
            "user_id": 'u6',
            "username": 'example_user_1',
            "is_active": True 
        },
        {
            "user_id": 'u7',
            "username": 'example_user_2',
            "is_active": True
        }
    ]
    print(team_add("ExampleTeam", members))
    print("\n")
    """

    print("=== Получаем список участников команд ===")
    print(team_get('Frontend'))
    print(team_get('Backend'))
    print(team_get('ExampleTeam'))
    print("\n")

    print("=== Устанавливаем статус active у пользоватлея u1 ===")
    print(user_set_is_active('u1', False))
    print("\n")

    print("=== Мерджим PR pr-1001 ===")
    print(pr_merge('pr-1001'))
    print("\n")

    print("\n")
    print("=== База данных пользователей после работы программы")
    print_db_users()
    print("\n")