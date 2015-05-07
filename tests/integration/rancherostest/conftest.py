import os
import pytest


@pytest.fixture(scope="session", autouse=True)
def my_own_session_run_at_beginning():
    os.chdir('../..')
    print('\nChdir to project root dir')
