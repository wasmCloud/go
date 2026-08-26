package blobstore

import (
	"io"
	"math"

	blobstore "go.wasmcloud.dev/component/imports/wasmcloud_blobstore_0_1_0_blobstore"
	container "go.wasmcloud.dev/component/imports/wasmcloud_blobstore_0_1_0_container"
	types "go.wasmcloud.dev/component/imports/wasmcloud_blobstore_0_1_0_types"
)

// ObjectID identifies an object together with its container.
type ObjectID struct {
	Container string
	Object    string
}

// ContainerMetadata is information about a container.
type ContainerMetadata struct {
	// Name is the container's name.
	Name string
	// CreatedAt is when the container was created, in nanoseconds since the
	// Unix epoch.
	CreatedAt uint64
}

// ObjectMetadata is information about an object.
type ObjectMetadata struct {
	// Name is the object's name.
	Name string
	// Container is the object's parent container.
	Container string
	// CreatedAt is when the object was created, in nanoseconds since the
	// Unix epoch.
	CreatedAt uint64
	// Size is the object's size in bytes.
	Size uint64
}

// Container is a handle to a collection of objects.
type Container struct {
	inner *container.Container
}

// CreateContainer creates a new empty container.
func CreateContainer(name string) (*Container, error) {
	res := blobstore.CreateContainer(name)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &Container{inner: res.Ok()}, nil
}

// GetContainer retrieves an existing container by name.
func GetContainer(name string) (*Container, error) {
	res := blobstore.GetContainer(name)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &Container{inner: res.Ok()}, nil
}

// DeleteContainer deletes a container and all objects within it.
func DeleteContainer(name string) error {
	res := blobstore.DeleteContainer(name)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// ContainerExists reports whether a container with the given name exists.
func ContainerExists(name string) (bool, error) {
	res := blobstore.ContainerExists(name)
	if res.IsErr() {
		return false, convertError(res.Err())
	}
	return res.Ok(), nil
}

// CopyObject duplicates an object, to the same or a different container,
// overwriting the destination if it exists.
func CopyObject(src, dest ObjectID) error {
	res := blobstore.CopyObject(src.toWit(), dest.toWit())
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// MoveObject moves or renames an object, to the same or a different
// container, overwriting the destination if it exists.
func MoveObject(src, dest ObjectID) error {
	res := blobstore.MoveObject(src.toWit(), dest.toWit())
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

func (id ObjectID) toWit() types.ObjectId {
	return types.ObjectId{Container: id.Container, Object: id.Object}
}

// Drop releases the host-side container handle. The handle is also released
// by the garbage collector if the Container becomes unreachable.
func (c *Container) Drop() {
	c.inner.Drop()
}

// Name returns the container's name.
func (c *Container) Name() (string, error) {
	res := c.inner.Name()
	if res.IsErr() {
		return "", convertError(res.Err())
	}
	return res.Ok(), nil
}

// Info returns the container's metadata.
func (c *Container) Info() (ContainerMetadata, error) {
	res := c.inner.Info()
	if res.IsErr() {
		return ContainerMetadata{}, convertError(res.Err())
	}
	md := res.Ok()
	return ContainerMetadata{Name: md.Name, CreatedAt: md.CreatedAt}, nil
}

// GetData retrieves the whole object as a byte stream. The caller must Close
// the returned reader (early Close cancels the transfer).
func (c *Container) GetData(name string) (io.ReadCloser, error) {
	return c.GetRange(name, 0, math.MaxUint64)
}

// GetRange retrieves a portion of an object as a byte stream. Start and end
// offsets are inclusive. The caller must Close the returned reader (early
// Close cancels the transfer).
func (c *Container) GetRange(name string, start, end uint64) (io.ReadCloser, error) {
	res := c.inner.GetData(name, start, end)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &streamReadCloser{stream: res.Ok()}, nil
}

// WriteData creates or replaces an object with the bytes read from data. The
// write completes once data is exhausted.
func (c *Container) WriteData(name string, data io.Reader) error {
	tx, rx := container.MakeStreamU8()
	go func() {
		defer tx.Drop()
		buf := make([]uint8, 16*1024)
		for !tx.ReaderDropped() {
			n, err := data.Read(buf)
			if n > 0 {
				tx.WriteAll(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()
	res := c.inner.WriteData(name, rx)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// ListObjects returns the names of the objects in the container, in no
// particular order.
func (c *Container) ListObjects() ([]string, error) {
	res := c.inner.ListObjects()
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	stream := res.Ok()
	defer stream.Drop()
	var names []string
	chunk := make([]string, 64)
	for {
		n := stream.Read(chunk)
		if n == 0 {
			return names, nil
		}
		names = append(names, chunk[:n]...)
	}
}

// DeleteObject deletes an object. Deleting a missing object is not an error.
func (c *Container) DeleteObject(name string) error {
	res := c.inner.DeleteObject(name)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// DeleteObjects deletes multiple objects in one call.
func (c *Container) DeleteObjects(names []string) error {
	res := c.inner.DeleteObjects(names)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// HasObject reports whether the object exists in this container.
func (c *Container) HasObject(name string) (bool, error) {
	res := c.inner.HasObject(name)
	if res.IsErr() {
		return false, convertError(res.Err())
	}
	return res.Ok(), nil
}

// ObjectInfo returns metadata for the object.
func (c *Container) ObjectInfo(name string) (ObjectMetadata, error) {
	res := c.inner.ObjectInfo(name)
	if res.IsErr() {
		return ObjectMetadata{}, convertError(res.Err())
	}
	md := res.Ok()
	return ObjectMetadata{
		Name:      md.Name,
		Container: md.Container,
		CreatedAt: md.CreatedAt,
		Size:      md.Size,
	}, nil
}

// Clear removes all objects within the container, leaving it empty.
func (c *Container) Clear() error {
	res := c.inner.Clear()
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}
