import pytest
import cattle
import requests
import json
from wait_for import wait_for


def headers(environment_id):
    return {
        'Accept': 'application/json',
        'x-api-project-id': environment_id
    }


DEFAULT_ENV = 'e1'
DEFAULT_HEADERS = headers(DEFAULT_ENV)
BASE_URL = 'http://localhost:8088/v1-catalog/'


def create_catalog(name, url, branch=None, headers=DEFAULT_HEADERS):
    schemas_url = 'http://localhost:8088/v1-catalog/schemas'
    client = cattle.from_env(url=schemas_url, headers=headers)

    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    data = {
        'name': name,
        'url': url,
    }
    if branch:
        data['branch'] = branch

    api_url = 'http://localhost:8088/v1-catalog/catalogs'
    response = requests.post(api_url, data=json.dumps(data), headers=headers)
    assert response.status_code == 200
    resp = response.json()
    assert resp['name'] == name
    assert resp['url'] == url
    if branch:
        assert resp['branch'] == branch

    api_url = 'http://localhost:8088/v1-catalog/templates?action=refresh'
    response = requests.post(api_url, headers=headers)
    assert response.status_code == 204

    templates = client.list_template()
    catalogs = client.list_catalog()
    assert len(catalogs) == len(original_catalogs) + 1
    assert len(templates) > len(original_templates)

    return resp


def create_duplicate_catalog(name, url, branch=None, headers=DEFAULT_HEADERS):
    schemas_url = 'http://localhost:8088/v1-catalog/schemas'
    client = cattle.from_env(url=schemas_url, headers=headers)

    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    data = {
        'name': name,
        'url': url,
    }
    if branch:
        data['branch'] = branch

    api_url = 'http://localhost:8088/v1-catalog/catalogs'
    response = requests.post(api_url, data=json.dumps(data), headers=headers)
    assert response.status_code == 422


def delete_catalog(name, headers=DEFAULT_HEADERS):
    schemas_url = 'http://localhost:8088/v1-catalog/schemas'
    client = cattle.from_env(url=schemas_url, headers=headers)

    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    url = 'http://localhost:8088/v1-catalog/catalogs/' + name
    response = requests.delete(url, headers=headers)
    assert response.status_code == 204

    templates = client.list_template()
    catalogs = client.list_catalog()
    assert len(catalogs) == len(original_catalogs) - 1
    assert len(templates) < len(original_templates)


@pytest.fixture
def client():
    url = 'http://localhost:8088/v1-catalog/schemas'
    catalogs = cattle.from_env(url=url, headers=DEFAULT_HEADERS).list_catalog()
    wait_for(
        lambda: len(catalogs) > 0
    )
    return cattle.from_env(url=url, headers=DEFAULT_HEADERS)


def test_catalog_list(client):
    catalogs = client.list_catalog()
    assert len(catalogs) == 2

    for catalog in catalogs:
        if catalog.name == 'orig':
            assert catalog.url == 'https://github.com/rancher/test-catalog'
        elif catalog.name == 'updated':
            assert catalog.url == '/tmp/test-catalog'
        else:
            assert False


def test_get_catalogs(client):
    url = 'http://localhost:8088/v1-catalog/catalogs'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    catalogs = response.json()['data']
    for c in catalogs:
        if c['name'] == 'orig':
            resp = c
            break
    assert resp['url'] == 'https://github.com/rancher/test-catalog'
    assert resp['links']['self'] == 'http://localhost:8088/' + \
        'v1-catalog/catalogs/orig?projectId=' + DEFAULT_ENV


def test_get_catalog(client):
    url = 'http://localhost:8088/v1-catalog/catalogs/orig'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['name'] == 'orig'
    assert resp['url'] == 'https://github.com/rancher/test-catalog'
    assert resp['links']['self'] == 'http://localhost:8088/' + \
        'v1-catalog/catalogs/orig?projectId=' + DEFAULT_ENV


def test_get_catalog_404(client):
    url = 'http://localhost:8088/v1-catalog/catalogs/not-real'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 404


def test_catalog_commit(client):
    latest_commit = '4ec17d4c057be16e01fecb599af16b2b9dda9065'
    url = 'http://localhost:8088/v1-catalog/catalogs/orig'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['commit'] == latest_commit

    url = 'http://localhost:8088/v1-catalog/catalogs/updated'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['commit'] != latest_commit


def test_create_and_delete_catalog(client):
    url = 'https://github.com/rancher/community-catalog'
    create_catalog('created', url)
    delete_catalog('created')


def test_catalog_branch(client):
    url = 'https://github.com/rancher/test-catalog'
    create_catalog('branch', url, "test-branch")
    delete_catalog('branch')


def test_catalog_duplicate_env_name(client):
    url = 'https://github.com/rancher/test-catalog'
    create_catalog('test', url)
    create_duplicate_catalog('test', url)
    delete_catalog('test')


def test_catalog_duplicate_global_name(client):
    # orig is the name of a global catalog that already exists
    url = 'https://github.com/rancher/test-catalog'
    create_duplicate_catalog('orig', url)


def test_catalog_edit(client):
    url = 'https://github.com/rancher/community-catalog'
    create_resp = create_catalog('edit', url)

    url = 'https://github.com/rancher/rancher-catalog'
    different_name = 'different_name'
    data = {
        'url': url,
        'name': different_name,
    }

    api_url = create_resp['links']['self']

    response = requests.put(api_url, data=json.dumps(data),
                            headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()

    response = requests.get(resp['links']['self'], headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()

    assert resp['url'] == url
    assert resp['name'] == different_name

    delete_catalog(different_name)


def test_catalog_different_environment(client):
    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    url = 'https://github.com/rancher/community-catalog'
    create_catalog('env', url, headers=headers('e2'))

    templates = client.list_template()
    catalogs = client.list_catalog()
    assert len(catalogs) == len(original_catalogs)
    assert len(templates) == len(original_templates)

    delete_catalog('env', headers=headers('e2'))


def test_template_list(client):
    templates = client.list_template()
    assert len(templates) > 0


def test_get_template(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:k8s'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()

    assert resp['catalogId'] == 'orig'
    assert resp['folderName'] == 'k8s'
    assert resp['defaultVersion'] == 'v1.3.0-rancher4'

    assert len(resp['categories']) == 1
    assert resp['categories'][0] == 'System'


def test_get_template_template_version(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:k8s-template-version'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()

    assert resp['catalogId'] == 'orig'
    assert resp['folderName'] == 'k8s-template-version'
    assert resp['defaultVersion'] == 'v1.3.0-rancher4'

    assert len(resp['categories']) == 1
    assert resp['categories'][0] == 'System'


def test_get_template_with_version_folders(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:version-folders'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()

    assert resp['catalogId'] == 'orig'
    assert resp['folderName'] == 'version-folders'

    versionLinks = resp['versionLinks']
    assert len(versionLinks) == 3
    assert 'v0.0.1' in versionLinks
    assert 'v0.0.1-rancher1.2' in versionLinks
    assert 'v0.0.3' in versionLinks

    for version in ('v0.0.1', 'v0.0.1-rancher1.2', 'v0.0.3'):
        version_id = 'orig:version-folders:' + version
        assert version_id in versionLinks.values()[0] or \
            version_id in versionLinks.values()[1] or \
            version_id in versionLinks.values()[2]


def test_get_template_404(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:not-real'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 404


def test_template_category(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:nfs-server'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['categories']) == 1
    assert resp['categories'][0] == 'Test'


def test_template_categories(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:categories'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['categories']) == 2
    assert resp['categories'][0] == 'category1'
    assert resp['categories'][1] == 'category2'


def test_preserve_category_case(client):
    url = BASE_URL + 'templates/orig:upper-case-categories'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['categories']) == 3
    assert resp['categories'][0] == 'CATEGORY1'
    assert resp['categories'][1] == 'CATEGORY2'
    assert resp['categories'][2] == 'CATEGORY3'


def test_category_filter(client):
    base_url = 'http://localhost:8088/v1-catalog/templates?category='
    for category in ('category1', 'category2', 'category3', 'System'):
        response = requests.get(base_url + category, headers=DEFAULT_HEADERS)
        assert response.status_code == 200
        resp = response.json()
        assert resp['data'] is not None

        for template in resp['data']:
            categories = [c.lower() for c in template['categories']]
            assert category.lower() in categories


def test_category_ne_filter(client):
    base_url = 'http://localhost:8088/v1-catalog/templates?category_ne='
    for category in ('category1', 'category2', 'System'):
        response = requests.get(base_url + category, headers=DEFAULT_HEADERS)
        assert response.status_code == 200
        resp = response.json()
        assert resp['data'] is not None

        for template in resp['data']:
            categories = template['categories']
            if categories:
                assert category not in template['categories']


def test_template_without_categories(client):
    base_url = 'http://localhost:8088/v1-catalog/templates'

    for category in ('category1', 'category2', 'System'):
        url = base_url + '?catalog_ne=' + category
        response = requests.get(url, headers=DEFAULT_HEADERS)
        assert response.status_code == 200
        resp = response.json()
        templates = resp['data']

        no_categories_template_found = False
        for template in templates:
            if template['folderName'] == 'no-categories':
                no_categories_template_found = True
                break

        assert no_categories_template_found


def test_machine_template(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:machine*vultr'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['templateBase'] == 'machine'

    url = 'http://localhost:8088/v1-catalog/templates/orig:machine*vultr:0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['files']) == 2
    assert 'rancher-compose.yml' in resp['files']
    assert 'url' in resp['files']


def test_template_labels(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:labels'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['labels'] is not None
    assert resp['labels']['key1'] == 'value1'
    assert resp['labels']['key2'] == 'value2'


def test_template_version_links(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:many-versions'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['versionLinks']) == 14

    url = 'http://localhost:8088/v1-catalog/templates/orig:many-versions' + \
        '?rancherVersion=v1.0.1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert len(resp['versionLinks']) == 9


def test_rancher_version_filter(client):
    templates = client.list_template()
    assert len(templates) > 0

    min_rancher_template_found = False
    max_rancher_template_found = False
    for template in templates:
        if template.folderName == 'min-rancher-version':
            min_rancher_template_found = True
        if template.folderName == 'max-rancher-version':
            max_rancher_template_found = True

    assert min_rancher_template_found
    assert max_rancher_template_found

    url = 'http://localhost:8088/v1-catalog/templates?rancherVersion=v1.2.0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['data'] is not None

    for template in resp['data']:
        assert template['folderName'] != 'min-rancher-version'

    url = 'http://localhost:8088/v1-catalog/templates?rancherVersion=v1.5.0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['data'] is not None

    for template in resp['data']:
        assert template['folderName'] != 'max-rancher-version'


def test_upgrade_links(client):
    url = 'http://localhost:8088/v1-catalog/templates/' + \
        'orig:test-upgrade-links:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    upgradeLinks = resp['upgradeVersionLinks']
    assert upgradeLinks is not None
    assert len(upgradeLinks) == 10

    url = 'http://localhost:8088/v1-catalog/templates/orig:many-versions:2' + \
        '?rancherVersion=v1.0.1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    upgradeLinks = resp['upgradeVersionLinks']
    assert upgradeLinks is not None
    assert len(upgradeLinks) == 7


def test_template_icon(client):

    url = 'http://localhost:8088/v1-catalog/templates/orig:nfs-server' + \
        '?image&projectId=%s' % (DEFAULT_ENV)
    response = requests.get(url, headers=headers(''))
    assert response.status_code == 200
    assert len(response.content) == 1139


def test_get_template_version_by_revision(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:k8s:0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == 'v0.1.0-rancher1'

    url = 'http://localhost:8088/v1-catalog/templates/orig:k8s:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == 'v1.2.4-rancher6'


def test_get_template_version_by_revision_template_version(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:k8s-template-version:0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == 'v0.1.0-rancher1'

    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:k8s-template-version:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == 'v1.2.4-rancher6'


def test_get_template_version_by_version(client):
    url = BASE_URL+'templates/orig:version-folders:v0.0.1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == 'v0.0.1'

    url = BASE_URL+'templates/orig:version-folders:v0.0.1-rancher1.2'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == 'v0.0.1-rancher1.2'

    url = BASE_URL+'templates/orig:version-folders:v0.0.3'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == 'v0.0.3'

    url = BASE_URL+'templates/orig:version-folders:v0.0.2'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 404


def test_get_template_version_404(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:k8s:1000'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 404


def test_get_template_version_404_template_version(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:k8s-template-version:1000'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 404


def test_get_template_version_labels(client):
    url = 'http://localhost:8088/v1-catalog/templates/orig:version-labels:0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['labels'] is not None
    assert resp['labels']['key1'] == 'value1'
    assert resp['labels']['key2'] == 'value2'


def test_template_version_questions(client):
    url = 'http://localhost:8088/v1-catalog/templates/' + \
        'orig:all-question-types:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    questions = resp['questions']
    assert questions is not None
    assert len(questions) == 11

    assert questions[0]['variable'] == 'TEST_STRING'
    assert questions[0]['label'] == 'String'
    assert not questions[0]['required']
    assert questions[0]['default'] == 'hello'
    assert questions[0]['type'] == 'string'

    assert questions[1]['variable'] == 'TEST_MULTILINE'
    assert questions[1]['label'] == 'Multi-Line'
    assert not questions[1]['required']
    assert questions[1]['default'] == 'Hello\nWorld\n'
    assert questions[1]['type'] == 'multiline'

    assert questions[2]['variable'] == 'TEST_PASSWORD'
    assert questions[2]['label'] == 'Password'
    assert not questions[2]['required']
    assert questions[2]['default'] == "not-so-secret stuff"
    assert questions[2]['type'] == 'password'

    assert questions[3]['variable'] == 'TEST_ENUM'
    assert questions[3]['label'] == 'Enum'
    assert not questions[3]['required']
    assert questions[3]['options'] == ['purple', 'monkey', 'dishwasher']
    assert questions[3]['default'] == 'monkey'
    assert questions[3]['type'] == 'enum'

    assert questions[4]['variable'] == 'TEST_DATE'
    assert questions[4]['label'] == 'Date'
    assert not questions[4]['required']
    assert questions[4]['default'] == '2015-07-25T19:55:00Z'
    assert questions[4]['type'] == 'date'

    assert questions[5]['variable'] == 'TEST_INT'
    assert questions[5]['label'] == 'Integer'
    assert not questions[5]['required']
    assert questions[5]['default'] == '42'
    assert questions[5]['type'] == 'int'

    assert questions[6]['variable'] == 'TEST_FLOAT'
    assert questions[6]['label'] == 'Float'
    assert not questions[6]['required']
    assert questions[6]['default'] == '4.2'
    assert questions[6]['type'] == 'float'

    assert questions[7]['variable'] == 'TEST_BOOLEAN'
    assert questions[7]['label'] == 'Boolean'
    assert not questions[7]['required']
    assert questions[7]['default'] == 'true'
    assert questions[7]['type'] == 'boolean'

    assert questions[8]['variable'] == 'TEST_SERVICE'
    assert questions[8]['label'] == 'Service'
    assert not questions[8]['required']
    assert questions[8]['default'] == 'kopf'
    assert questions[8]['type'] == 'service'

    assert questions[9]['variable'] == 'TEST_CERTIFICATE'
    assert questions[9]['label'] == 'Certificate'
    assert not questions[9]['required']
    assert questions[9]['default'] == 'rancher.rocks'
    assert questions[9]['type'] == 'certificate'

    assert questions[10]['variable'] == 'TEST_UNKNOWN'
    assert questions[10]['label'] == 'Unknown'
    assert not questions[10]['required']
    assert questions[10]['default'] == 'wha?'
    assert questions[10]['type'] == 'unknown'


def test_template_version_questions_template_version(client):
    url = 'http://localhost:8088/v1-catalog/templates/' + \
        'orig:all-question-types-template-version:1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    questions = resp['questions']
    assert questions is not None
    assert len(questions) == 11

    assert questions[0]['variable'] == 'TEST_STRING'
    assert questions[0]['label'] == 'String'
    assert not questions[0]['required']
    assert questions[0]['default'] == 'hello'
    assert questions[0]['type'] == 'string'

    assert questions[1]['variable'] == 'TEST_MULTILINE'
    assert questions[1]['label'] == 'Multi-Line'
    assert not questions[1]['required']
    assert questions[1]['default'] == 'Hello\nWorld\n'
    assert questions[1]['type'] == 'multiline'

    assert questions[2]['variable'] == 'TEST_PASSWORD'
    assert questions[2]['label'] == 'Password'
    assert not questions[2]['required']
    assert questions[2]['default'] == "not-so-secret stuff"
    assert questions[2]['type'] == 'password'

    assert questions[3]['variable'] == 'TEST_ENUM'
    assert questions[3]['label'] == 'Enum'
    assert not questions[3]['required']
    assert questions[3]['options'] == ['purple', 'monkey', 'dishwasher']
    assert questions[3]['default'] == 'monkey'
    assert questions[3]['type'] == 'enum'

    assert questions[4]['variable'] == 'TEST_DATE'
    assert questions[4]['label'] == 'Date'
    assert not questions[4]['required']
    assert questions[4]['default'] == '2015-07-25T19:55:00Z'
    assert questions[4]['type'] == 'date'

    assert questions[5]['variable'] == 'TEST_INT'
    assert questions[5]['label'] == 'Integer'
    assert not questions[5]['required']
    assert questions[5]['default'] == '42'
    assert questions[5]['type'] == 'int'

    assert questions[6]['variable'] == 'TEST_FLOAT'
    assert questions[6]['label'] == 'Float'
    assert not questions[6]['required']
    assert questions[6]['default'] == '4.2'
    assert questions[6]['type'] == 'float'

    assert questions[7]['variable'] == 'TEST_BOOLEAN'
    assert questions[7]['label'] == 'Boolean'
    assert not questions[7]['required']
    assert questions[7]['default'] == 'true'
    assert questions[7]['type'] == 'boolean'

    assert questions[8]['variable'] == 'TEST_SERVICE'
    assert questions[8]['label'] == 'Service'
    assert not questions[8]['required']
    assert questions[8]['default'] == 'kopf'
    assert questions[8]['type'] == 'service'

    assert questions[9]['variable'] == 'TEST_CERTIFICATE'
    assert questions[9]['label'] == 'Certificate'
    assert not questions[9]['required']
    assert questions[9]['default'] == 'rancher.rocks'
    assert questions[9]['type'] == 'certificate'

    assert questions[10]['variable'] == 'TEST_UNKNOWN'
    assert questions[10]['label'] == 'Unknown'
    assert not questions[10]['required']
    assert questions[10]['default'] == 'wha?'
    assert questions[10]['type'] == 'unknown'


def test_refresh(client):
    url = 'http://localhost:8088/v1-catalog/templates/updated:many-versions:14'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['version'] == '1.0.14'


def test_refresh_no_changes(client):
    original_catalogs = client.list_catalog()
    assert len(original_catalogs) > 0
    original_templates = client.list_template()
    assert len(original_templates) > 0

    url = 'http://localhost:8088/v1-catalog/templates?action=refresh'
    response = requests.post(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 204

    catalogs = client.list_catalog()
    templates = client.list_template()
    assert len(catalogs) == len(original_catalogs)
    assert len(templates) == len(original_templates)


def test_v2_syntax(client):
    for revision in [0, 1, 2, 3]:
        url = 'http://localhost:8088/v1-catalog/templates/orig:v2:' + \
            str(revision)
        response = requests.get(url, headers=DEFAULT_HEADERS)
        assert response.status_code == 200


def test_alternative_config_fields_1(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:alternative-config-fields-1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['defaultVersion'] == '3.0.0'
    assert resp['links']['project'] == 'www.test.com'


def test_alternative_config_fields_2(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:alternative-config-fields-2'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['defaultVersion'] == '3.0.0'
    assert resp['links']['project'] == 'www.test.com'


def test_alternative_config_fields_3(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:alternative-config-fields-3'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['defaultVersion'] == '3.0.0'


def test_default_versions_1(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:default-versions-1'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()

    assert resp['links']['defaultVersion'] == 'http://' + \
        'localhost:8088/v1-catalog/templates/orig:default-versions-1:2'

    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:default-versions-1:0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['links']['defaultUpgradeVersion'] == 'http://localhost:' + \
        '8088/v1-catalog/templates/orig:default-versions-1:2'


def test_default_versions_2(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:default-versions-2'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['links']['defaultVersion'] == 'http://' + \
        'localhost:8088/v1-catalog/templates/orig:default-versions-2:3'

    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:default-versions-2:0'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()

    assert resp['links']['defaultUpgradeVersion'] == 'http://' + \
        'localhost:8088/v1-catalog/templates/orig:default-versions-2:2'


def test_default_versions_3(client):
    url = 'http://localhost:8088/v1-catalog/templates' + \
        '/orig:default-versions-3'
    response = requests.get(url, headers=DEFAULT_HEADERS)
    assert response.status_code == 200
    resp = response.json()
    assert resp['links']['defaultVersion'] == 'http://' + \
        'localhost:8088/v1-catalog/templates/orig:default-versions-3:2'
