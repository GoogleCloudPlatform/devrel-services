from dataclasses import dataclass
from dateutil import parser
from datetime import datetime
from datetime import timedelta
import json
import os
import sys
import subprocess
from typing import List

from google.cloud import storage
import requests


DRGHS_API_KEY = os.getenv('DRGHS_API_KEY')
GITHUB_API_KEY = os.getenv('GITHUB_API_KEY')

BUCKET_NAME = 'devrel-prod-settings'
BLOB_NAME = 'public_repos.json'

SECONDS_IN_AN_HOUR = 60*60


@dataclass
class Commit:
  sha: str
  author_time: datetime


@dataclass
class Repo:
    owner: str
    repository: str
    branch: str


@dataclass
class ProblemRepo:
    repo: Repo
    delta: timedelta


def get_tracked_repos() -> List[Repo]:
    storage_client = storage.Client()
    bucket = storage_client.bucket(BUCKET_NAME)
    blob = bucket.blob(BLOB_NAME)

    content = blob.download_as_bytes().decode('utf-8')
    data = json.loads(content)
    repos = data.get('repos', [])

    retval = []

    for repo in repos:
        if not repo['is_tracking_samples']:
            continue
        # 'repo' field is of form 'owner/repository'
        orp = repo.get('repo', '').split('/')
        # If default_branch is unset, use 'master'
        retval.append(Repo(orp[0], orp[1], repo.get('default_branch', 'master')))

    return retval


def get_latest_samplr_commit_for_repo(repo: Repo) -> Commit:
    npt = ''
    keepon = True
    last: Commit = None
    while keepon:
        resp = requests.get(f'https://samplr.endpoints.devrel-prod.cloud.goog/v1/owners/{repo.owner}/repositories/{repo.repository}/gitCommits?key={DRGHS_API_KEY}&page_token={npt}')
        data = resp.json()
        npt = data.get('nextPageToken', '')
        commits = data.get('gitCommits', [])
        if not commits:
            print(f'Got no commits for repo: {repo}')
            return None
        lastcommit = commits[-1]
        last = Commit(lastcommit['sha'], parser.isoparse(lastcommit['committedTime']))
        if not npt:
            keepon = False
    return last


def get_latest_github_commit_for_repo(repo: Repo) -> Commit:
    headers = {'Authorization': f'Bearer {GITHUB_API_KEY}'}
    resp = requests.get(
        f'https://api.github.com/repos/{repo.owner}/{repo.repository}/branches/{repo.branch}',
        headers=headers)
    data = resp.json()
    return Commit(
        data['commit']['sha'],
        parser.isoparse(data['commit']['commit']['committer']['date']))


def scale_repo(repo: Repo):
    result = subprocess.run(
        [
            'kubectl', 'get', 'deployment', '-l',
            f'samplr-sprvsr-autogen=true,owner={repo.owner},repository={repo.repository}',
            '-o', 'jsonpath="{.items[0].metadata.name}"'
        ],
        capture_output=True)
    deployment_name = result.stdout.decode('utf-8').replace('"','')


def main():

    if not DRGHS_API_KEY:
        print('Must set environment variable DRGHS_API_KEY')
        sys.exit(1)
    if not GITHUB_API_KEY:
        print('Must set environment variable GITHUB_API_KEY')
        sys.exit(1)


    problems = []
    repos = get_tracked_repos()

    for repo in repos:
        github_commit = get_latest_github_commit_for_repo(repo)
        samplr_commit = get_latest_samplr_commit_for_repo(repo)

        if not github_commit:
            print(f'Got no github commit for {repo}')
            continue
        if not samplr_commit:
            print(f'Got no samplr commit for {repo}')
            continue

        time_delta = (github_commit.author_time - samplr_commit.author_time)
        if samplr_commit.sha != github_commit.sha and time_delta.total_seconds() > SECONDS_IN_AN_HOUR:
            problems.append(ProblemRepo(repo, time_delta))
            print(f'Problem Repo: {repo}. It has been {time_delta} since it was updated')


    print(f'{len(problems)}/{len(repos)} Were Problems')

if __name__ == '__main__':
    main()
