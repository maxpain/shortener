package utils

import "net/url"

func СonstructUrl(base, postfix string) (string, error) {
	baseURL, err := url.Parse(base)

	if err != nil {
		return "", err
	}

	relativeURL, err := url.Parse(postfix)

	if err != nil {
		return "", err
	}

	fullURL := baseURL.ResolveReference(relativeURL)
	return fullURL.String(), nil
}
