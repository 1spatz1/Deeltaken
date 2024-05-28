package utility

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var cli, _ = client.NewClientWithOpts(client.FromEnv)

func RunContainer(imageName string, isLocalImage bool, code string) (string, error) {
	// Pull the image
	if !isLocalImage {
		if err := PullImage(imageName); err != nil {
			return "", err
		}
	}

	// Check if container is already running
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		println("Error listing containers", err.Error())
		return "", err
	}
	println("Found", len(containers), "containers")

	containerID := ""
	for _, container := range containers {
		if strings.Split(container.Image, ":")[0] == strings.Split(imageName, ":")[0] {
			println("Found container with image: " + imageName + " with ID: " + container.ID)
			containerID = container.ID
			break
		}
	}

	if containerID == "" {
		// Run a image same as command: `docker run -v "ABSINFILELOC:/input.txt IMAGENAME`
		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image: imageName,
		}, &container.HostConfig{
			AutoRemove: true,
		}, nil, nil, "")
		if err != nil {
			println("Error creating container", err.Error())
			return "", err
		}

		containerID = resp.ID

		// Start the container
		if err := cli.ContainerStart(context.Background(), containerID, container.StartOptions{}); err != nil {
			println("Error starting container", err.Error())
			return "", err
		}
	}

	// Run this command: docker exec -i container_id2 sh -c 'cat > ./bar/foo.txt' < ./input.txt
	cmd := exec.Command("docker", "exec", "-i", containerID, "sh", "-c", "cat > /input.txt")

	// Create a pipe to the stdin of the command
	stdin, err := cmd.StdinPipe()
	if err != nil {
		println("Error creating stdin pipe", err.Error())
		return "", err
	}

	// Write the code to the stdin of the command
	if _, err := stdin.Write([]byte(code)); err != nil {
		println("Error writing to stdin", err.Error())
		return "", err
	}

	// Close the stdin pipe
	if err := stdin.Close(); err != nil {
		println("Error closing stdin", err.Error())
		return "", err
	}

	if err := cmd.Run(); err != nil {
		println("Error running command", err.Error())
		return "", err
	}

	cmd = exec.Command("docker", "exec", containerID, "/source/script.sh")

	// Get the stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		println("Error creating stdout pipe", err.Error())
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		println("Error creating stderr pipe", err.Error())
		return "", err
	}

	// Start the command before setting up stream reading
	if err := cmd.Start(); err != nil {
		println("Error starting command", err.Error())
		return "", err
	}

	// Create a channel to collect the lines
	lines := make(chan string)

	// Create a WaitGroup to wait for the goroutines to finish
	var wg sync.WaitGroup
	wg.Add(2) // We have two goroutines

	// Read stdout line by line
	go func() {
		defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println("stdout:", scanner.Text()) // Print each line of stdout
			lines <- scanner.Text()
		}
		println("finished stdout")
	}()

	// Read stderr line by line
	go func() {
		defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println("stderr:", scanner.Text()) // Print each line of stderr
			lines <- scanner.Text()
		}
		println("finished stderr")
	}()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		println("Error waiting for command", err.Error())
		return "", err
	}
	
	// Wait for the goroutines to finish
	wg.Wait()
	println("passed wait")
	// Now it's safe to close the lines channel
	close(lines)

	// Collect all lines into a single string
	output := ""
	for line := range lines {
		output += line + "\n"
	}

	// Remove the header from the log output
	if len(output) < 1 {
		return "", fmt.Errorf("error: No output")
	}
	output = strings.Trim(output, "\n")

	if strings.HasPrefix(output, "Error") {
		return "", fmt.Errorf(output)
	}

	return output, nil
}

func IsImagePresent(imageName string) bool {
	list, err := cli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, ims := range list {
		for _, tag := range ims.RepoTags {
			if strings.HasPrefix(tag, imageName) {
				return true
			}
		}
	}

	return false
}

func PullImage(imageName string) error {
	// Always pull the image to make sure it is up to date
	_, err := cli.ImagePull(context.Background(), imageName, image.PullOptions{})

	// Wait for any image to be usable
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		for !IsImagePresent(imageName) {
			time.Sleep(1 * time.Second)
		}
		wg.Done()
	}(&wg)
	wg.Wait()

	return err
}
