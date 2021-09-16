package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	for _, fileName := range os.Args[1:] {
		if err := processFile(fileName); err != nil {
			log.Fatal(err)
		}
	}
}

func processFile(fileName string) error {
	in, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	doc := struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name      string `yaml:"name"`
			Namespace string `yaml:"namespace"`
		} `yaml:"metadata"`
		Data struct {
			Example string `yaml:"_example"`
		}
	}{}
	if err := yaml.Unmarshal(in, &doc); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	var exampleNode yaml.Node
	if err := yaml.Unmarshal([]byte(doc.Data.Example), &exampleNode); err != nil {
		return fmt.Errorf("failed to parse _example YAML: %w", err)
	}

	name := doc.Metadata.Name[strings.Index(doc.Metadata.Name, "-")+1:]
	name = strings.ToTitle(name[0:1]) + name[1:]
	fmt.Printf("# Configuring the %s ConfigMap\n", name)

	fmt.Printf("You can view the current `%s` ConfigMap by running the following command:\n", doc.Metadata.Name)
	fmt.Println("```yaml")
	fmt.Printf("kubectl get configmap -n %s %s -oyaml\n", doc.Metadata.Namespace, doc.Metadata.Name)
	fmt.Println("```")

	fmt.Printf("## Example %s ConfigMap\n", doc.Metadata.Name)
	fmt.Println("```yaml")
	fmt.Println("apiVersion: ", doc.APIVersion)
	fmt.Println("kind: ", doc.Kind)
	fmt.Println("metadata:")
	fmt.Println("  name: ", doc.Metadata.Name)
	fmt.Println("  namespace: ", doc.Metadata.Namespace)
	fmt.Println("data:")
	for i := 0; i < len(exampleNode.Content[0].Content); i += 2 {
		n := exampleNode.Content[0].Content[i]
		v := exampleNode.Content[0].Content[i+1]
		fmt.Print("  ", n.Value)
		fmt.Print(": ")
		b, err := yaml.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Print(string(b))
	}
	fmt.Println("```")
	fmt.Println("See below for a description of each property.")
	fmt.Println("## Properties")

	for i := 0; i < len(exampleNode.Content[0].Content); i += 2 {
		n := exampleNode.Content[0].Content[i]
		v := exampleNode.Content[0].Content[i+1]
		b, err := yaml.Marshal(v)
		if err != nil {
			return err
		}

		words := strings.Split(n.Value, "-")
		for j, w := range words {
			words[j] = strings.ToTitle(w[0:1]) + w[1:]
		}
		fmt.Println("###", strings.Join(words, " "))

		lines := strings.Split(n.HeadComment, "\n")
		fmt.Println("{% raw %}")
		for _, l := range lines {
			fmt.Println(strings.TrimSpace(strings.TrimPrefix(l, "#")))
		}
		fmt.Println("{% endraw %}")

		fmt.Println()
		fmt.Printf("**Key**: `%s`\n\n", n.Value)
		fmt.Printf("**Default**: `%s`\n\n", string(b))
		fmt.Println()
		fmt.Println("**Example:**")
		fmt.Println("```yaml")
		fmt.Println("apiVersion: ", doc.APIVersion)
		fmt.Println("kind: ", doc.Kind)
		fmt.Println("metadata:")
		fmt.Println("  name: ", doc.Metadata.Name)
		fmt.Println("  namespace: ", doc.Metadata.Namespace)
		fmt.Println("data:")
		fmt.Print("  ", n.Value)
		fmt.Print(": ")
		fmt.Println(string(b))
		fmt.Println("```")
		fmt.Println()
	}

	return nil
}
