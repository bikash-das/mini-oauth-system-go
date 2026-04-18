package utils

import "net/url"

func BuildURL(base string, params map[string]string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
