import subprocess

import os
import pytest
import rostest


@pytest.fixture(scope="session", autouse=True)
def chdir_to_project_root():
    os.chdir(os.path.join(os.path.dirname(rostest.__file__), '../../..'))
    print('\nChdir to project root dir: ' + subprocess.check_output('pwd'))
    os.chmod('./tests/integration/assets/test.key', 0o600)
    print('Also, `chmod 600 tests/integration/assets/test.key` to make ssh happy')
