This project is a simple webhook tester implemented using Golang and developed following the principles of Test-Driven Development (TDD). It provides a convenient way to test your webhook program without the need for a public webhook link. The project is completely automated using CI/CD with GitHub Actions, ensuring that the code is properly tested and built with each change.

If you have a webhook program that you want to test but don't have a publicly accessible webhook URL, this webhook tester can come in handy. It allows you to simulate receiving webhook requests.

Feel free to explore the project, contribute to its development, and leverage it for your own webhook testing needs. The project is open-source and welcomes contributions from the community to make it even better.

### How it works:

when you run the client executable  file it makes a websocket connection with the server, the server generates a random URL and sends it to the client, the client displays it in the console,

![client](https://github.com/sanjayJ369/webhook-tester/assets/111266023/53dc2b0f-3744-42df-8042-e77fab97a3d7)

here the webhook program that i need to test is running on the `localhost:5555`, the link displayed is used as the webhook link to receive webhooks(HTTP POST request), when a post request is sent to the provided link, the server receives it, then the server serialises the request and sends it to the client running on your system, client on receiving the serialized request rebuilds the request and makes the same request to the webhook program running on `localhost:5555`

[**Read My Blog To Know More**](https://blog.sanjayj.dev/blog/webhook-tester/)

## How to use:

### **Using Binary Files**

[**Download latest file**](https://github.com/sanjayJ369/webhook-tester/tree/main/binary-files)

To use the Simple Webhook Tester with binary files, follow these steps:

1. Download the executable file corresponding to your operating system and architecture.
2. If you are using Linux, ensure that the file has executable permissions. You can do this by running the command **`chmod +x whclient-linux-amd64`**.
3. Run the executable file from the terminal, specifying the port on which your webhook program is running. For example: **`./whclient-linux-amd64 5555`**

This will start the webhook client on port 5555.

### **Using Go**

Alternatively, if you have Go version 1.22.2 or later installed on your computer, you can use the following steps:

1. Clone this repository.
2. Change into the repository directory.
3. Run the command **`go run cmd/client/main.go 5555`** to start the webhook client on port 5555.

Note: Replace

```
5555
```

with the actual port number on which your webhook program is running.
