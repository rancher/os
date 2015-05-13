import os
import pytest


@pytest.fixture(scope="session", autouse=True)
def chdir_to_project_root():
    os.chdir('../..')
    print('\nChdir to project root dir')
    os.chmod('./tests/integration/assets/test.key', 0o600)
    print('Also, `chmod 600 tests/integration/assets/test.key` to make ssh happy')
