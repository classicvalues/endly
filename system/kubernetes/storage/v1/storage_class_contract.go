package v1

import (
	"errors"
	"fmt"
	vvc "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/typed/storage/v1"
)

/*autogenerated contract adapter*/

//StorageClassCreateRequest represents request
type StorageClassCreateRequest struct {
	service_ v1.StorageClassInterface
	*vvc.StorageClass
}

//StorageClassUpdateRequest represents request
type StorageClassUpdateRequest struct {
	service_ v1.StorageClassInterface
	*vvc.StorageClass
}

//StorageClassDeleteRequest represents request
type StorageClassDeleteRequest struct {
	service_ v1.StorageClassInterface
	Name     string
	*metav1.DeleteOptions
}

//StorageClassDeleteCollectionRequest represents request
type StorageClassDeleteCollectionRequest struct {
	service_ v1.StorageClassInterface
	*metav1.DeleteOptions
	ListOptions metav1.ListOptions
}

//StorageClassGetRequest represents request
type StorageClassGetRequest struct {
	service_ v1.StorageClassInterface
	Name     string
	metav1.GetOptions
}

//StorageClassListRequest represents request
type StorageClassListRequest struct {
	service_ v1.StorageClassInterface
	metav1.ListOptions
}

//StorageClassWatchRequest represents request
type StorageClassWatchRequest struct {
	service_ v1.StorageClassInterface
	metav1.ListOptions
}

//StorageClassPatchRequest represents request
type StorageClassPatchRequest struct {
	service_     v1.StorageClassInterface
	Name         string
	Pt           types.PatchType
	Data         []byte
	Subresources []string
}

func init() {
	register(&StorageClassCreateRequest{})
	register(&StorageClassUpdateRequest{})
	register(&StorageClassDeleteRequest{})
	register(&StorageClassDeleteCollectionRequest{})
	register(&StorageClassGetRequest{})
	register(&StorageClassListRequest{})
	register(&StorageClassWatchRequest{})
	register(&StorageClassPatchRequest{})
}

func (r *StorageClassCreateRequest) SetService(service interface{}) error {
	var ok bool
	if r.service_, ok = service.(v1.StorageClassInterface); !ok {
		return fmt.Errorf("invalid service type: %T, expected: v1.StorageClassInterface", service)
	}
	return nil
}

func (r *StorageClassCreateRequest) Call() (result interface{}, err error) {
	if r.service_ == nil {
		return nil, errors.New("service was empty")
	}
	result, err = r.service_.Create(r.StorageClass)
	return result, err
}

func (r *StorageClassCreateRequest) GetId() string {
	return "storage/v1.StorageClass.Create"
}

func (r *StorageClassUpdateRequest) SetService(service interface{}) error {
	var ok bool
	if r.service_, ok = service.(v1.StorageClassInterface); !ok {
		return fmt.Errorf("invalid service type: %T, expected: v1.StorageClassInterface", service)
	}
	return nil
}

func (r *StorageClassUpdateRequest) Call() (result interface{}, err error) {
	if r.service_ == nil {
		return nil, errors.New("service was empty")
	}
	result, err = r.service_.Update(r.StorageClass)
	return result, err
}

func (r *StorageClassUpdateRequest) GetId() string {
	return "storage/v1.StorageClass.Update"
}

func (r *StorageClassDeleteRequest) SetService(service interface{}) error {
	var ok bool
	if r.service_, ok = service.(v1.StorageClassInterface); !ok {
		return fmt.Errorf("invalid service type: %T, expected: v1.StorageClassInterface", service)
	}
	return nil
}

func (r *StorageClassDeleteRequest) Call() (result interface{}, err error) {
	if r.service_ == nil {
		return nil, errors.New("service was empty")
	}
	err = r.service_.Delete(r.Name, r.DeleteOptions)
	return result, err
}

func (r *StorageClassDeleteRequest) GetId() string {
	return "storage/v1.StorageClass.Delete"
}

func (r *StorageClassDeleteCollectionRequest) SetService(service interface{}) error {
	var ok bool
	if r.service_, ok = service.(v1.StorageClassInterface); !ok {
		return fmt.Errorf("invalid service type: %T, expected: v1.StorageClassInterface", service)
	}
	return nil
}

func (r *StorageClassDeleteCollectionRequest) Call() (result interface{}, err error) {
	if r.service_ == nil {
		return nil, errors.New("service was empty")
	}
	err = r.service_.DeleteCollection(r.DeleteOptions, r.ListOptions)
	return result, err
}

func (r *StorageClassDeleteCollectionRequest) GetId() string {
	return "storage/v1.StorageClass.DeleteCollection"
}

func (r *StorageClassGetRequest) SetService(service interface{}) error {
	var ok bool
	if r.service_, ok = service.(v1.StorageClassInterface); !ok {
		return fmt.Errorf("invalid service type: %T, expected: v1.StorageClassInterface", service)
	}
	return nil
}

func (r *StorageClassGetRequest) Call() (result interface{}, err error) {
	if r.service_ == nil {
		return nil, errors.New("service was empty")
	}
	result, err = r.service_.Get(r.Name, r.GetOptions)
	return result, err
}

func (r *StorageClassGetRequest) GetId() string {
	return "storage/v1.StorageClass.Get"
}

func (r *StorageClassListRequest) SetService(service interface{}) error {
	var ok bool
	if r.service_, ok = service.(v1.StorageClassInterface); !ok {
		return fmt.Errorf("invalid service type: %T, expected: v1.StorageClassInterface", service)
	}
	return nil
}

func (r *StorageClassListRequest) Call() (result interface{}, err error) {
	if r.service_ == nil {
		return nil, errors.New("service was empty")
	}
	result, err = r.service_.List(r.ListOptions)
	return result, err
}

func (r *StorageClassListRequest) GetId() string {
	return "storage/v1.StorageClass.List"
}

func (r *StorageClassWatchRequest) SetService(service interface{}) error {
	var ok bool
	if r.service_, ok = service.(v1.StorageClassInterface); !ok {
		return fmt.Errorf("invalid service type: %T, expected: v1.StorageClassInterface", service)
	}
	return nil
}

func (r *StorageClassWatchRequest) Call() (result interface{}, err error) {
	if r.service_ == nil {
		return nil, errors.New("service was empty")
	}
	result, err = r.service_.Watch(r.ListOptions)
	return result, err
}

func (r *StorageClassWatchRequest) GetId() string {
	return "storage/v1.StorageClass.Watch"
}

func (r *StorageClassPatchRequest) SetService(service interface{}) error {
	var ok bool
	if r.service_, ok = service.(v1.StorageClassInterface); !ok {
		return fmt.Errorf("invalid service type: %T, expected: v1.StorageClassInterface", service)
	}
	return nil
}

func (r *StorageClassPatchRequest) Call() (result interface{}, err error) {
	if r.service_ == nil {
		return nil, errors.New("service was empty")
	}
	result, err = r.service_.Patch(r.Name, r.Pt, r.Data, r.Subresources...)
	return result, err
}

func (r *StorageClassPatchRequest) GetId() string {
	return "storage/v1.StorageClass.Patch"
}
