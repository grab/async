package async

import "context"

func upload(ctx context.Context, file string) (string, error) {
	// do file uploading
	return "", nil
}

func uploadFilesConcurrently(files []string) {
	var tasks []Task[string]
	for _, file := range files {
		f := file

		tasks = append(
			tasks, NewTask(
				func(ctx context.Context) (string, error) {
					return upload(ctx, f)
				},
			),
		)
	}

	ForkJoin(context.Background(), tasks)
}

func main() {

}
