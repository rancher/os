import os
import pytest


@pytest.fixture(scope="session", autouse=True)
def chdir_to_project_root():
    os.chdir('../..')
    print('\nChdir to project root dir')
