package amphora

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imagedata"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

func (b *bootstrap) EnsureImage(ctx context.Context) error {
	imageURL := b.instance.Spec.Amphora.ImageURL

	image, err := b.getCurrentImage(imageURL)
	if err != nil {
		return err
	} else if image == nil {
		image, err = b.uploadImage(imageURL)
		if err != nil {
			return err
		}
	}

	if b.instance.Status.Amphora.ImageProjectID == image.Owner {
		return nil
	}

	b.instance.Status.Amphora.ImageProjectID = image.Owner
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}
	return nil
}

func (b *bootstrap) getCurrentImage(imageURL string) (*images.Image, error) {
	pager := images.List(b.image, images.ListOpts{
		Tags: []string{amphoraImageTag},
	})

	page, err := pager.AllPages()
	if err != nil {
		return nil, err
	}

	images, err := images.ExtractImages(page)
	if err != nil {
		return nil, err
	}

	for _, image := range images {
		if image.Properties[imageSourceProperty] == imageURL {
			return &image, nil
		}
	}

	return nil, nil
}

func (b *bootstrap) uploadImage(imageURL string) (*images.Image, error) {
	b.log.Info("Creating image", "name", amphoraImageName)
	image, err := images.Create(b.image, images.CreateOpts{
		Name:            amphoraImageName,
		Tags:            []string{amphoraImageTag},
		ContainerFormat: "bare",
		DiskFormat:      "qcow2",
		Properties: map[string]string{
			imageSourceProperty: imageURL,
		},
	}).Extract()
	if err != nil {
		return nil, err
	}

	b.log.Info("Uploading image",
		"name", image.Name,
		"url", imageURL)
	data, err := fetchImage(imageURL)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	if err := imagedata.Upload(b.image, image.ID, data).Err; err != nil {
		return nil, err
	}

	return image, nil
}

func fetchImage(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code for %s: %d", url, resp.StatusCode)
	}

	return resp.Body, nil
}
