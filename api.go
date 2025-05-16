package main

type result struct {
	code int
	msg  string
}

type recvTopicBody struct {
	Topic
	Post
}

// recv json like
//
//	{
//		"title": "test",
//		"model_id": 1,
//		"content": "Hello World!"
//	}
