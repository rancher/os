package helm

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/rancher/catalog-service/model"
	"gopkg.in/yaml.v2"
)

func DownloadIndex(indexURL string) (*HelmRepoIndex, error) {
	if indexURL[len(indexURL)-1:] == "/" {
		indexURL = indexURL[:len(indexURL)-1]
	}
	indexURL = indexURL + "/index.yaml"
	resp, err := http.Get(indexURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	sum := md5.Sum(body)
	hash := hex.EncodeToString(sum[:])

	helmRepoIndex := &HelmRepoIndex{
		URL:       indexURL,
		IndexFile: &IndexFile{},
		Hash:      hash,
	}
	return helmRepoIndex, yaml.Unmarshal(body, helmRepoIndex.IndexFile)
}

func SaveIndex(index *HelmRepoIndex, repoPath string) error {
	fileBytes, err := yaml.Marshal(index.IndexFile)
	if err != nil {
		return err
	}

	indexPath := path.Join(repoPath, "index.yaml")

	f, err := os.OpenFile(indexPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return nil
	}

	_, err = f.Write(fileBytes)
	return err
}

func LoadIndex(repoPath string) (*HelmRepoIndex, error) {
	indexPath := path.Join(repoPath, "index.yaml")

	f, err := os.Open(indexPath)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	sum := md5.Sum(body)
	hash := hex.EncodeToString(sum[:])

	helmRepoIndex := &HelmRepoIndex{
		IndexFile: &IndexFile{},
		Hash:      hash,
	}
	return helmRepoIndex, yaml.Unmarshal(body, helmRepoIndex.IndexFile)
}

func FetchFiles(urls []string) ([]model.File, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	files := []model.File{}
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		gzf, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gzf.Close()

		tarReader := tar.NewReader(gzf)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, err
			}

			switch header.Typeflag {
			case tar.TypeDir:
				continue
			case tar.TypeReg:
				fallthrough
			case tar.TypeRegA:
				name := header.Name
				contents, err := ioutil.ReadAll(tarReader)
				if err != nil {
					return nil, err
				}
				files = append(files, filterFile(model.File{
					Name:     name,
					Contents: string(contents),
				}))
			}
		}
	}
	return files, nil
}

func LoadMetadata(path string) (*ChartMetadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	metadata := &ChartMetadata{}
	return metadata, yaml.Unmarshal(data, metadata)
}

func LoadFile(path string) (*model.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	filteredFile := filterFile(model.File{
		Name:     f.Name(),
		Contents: string(data),
	})
	return &filteredFile, nil
}
