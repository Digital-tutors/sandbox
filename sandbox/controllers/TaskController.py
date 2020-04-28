import asyncio
import aiohttp
from sandbox.models.Task import Task


async def get_task(url):
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as response:
            return response.json()


def get_task_by_url(url):
    loop = asyncio.get_event_loop()
    result = loop.run_until_complete(asyncio.wait(get_task(url)))
    task = Task(
        task_id=result["id"],
        topic_id=result["topicId"],
        author_id=result["authorId"],
        description=result["description"],
        options=result["options"],
        tests=result["tests"])
    return task
