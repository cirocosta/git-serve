//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Addressable) DeepCopyInto(out *Addressable) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Addressable.
func (in *Addressable) DeepCopy() *Addressable {
	if in == nil {
		return nil
	}
	out := new(Addressable)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServer) DeepCopyInto(out *GitServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServer.
func (in *GitServer) DeepCopy() *GitServer {
	if in == nil {
		return nil
	}
	out := new(GitServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerList) DeepCopyInto(out *GitServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GitServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerList.
func (in *GitServerList) DeepCopy() *GitServerList {
	if in == nil {
		return nil
	}
	out := new(GitServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpec) DeepCopyInto(out *GitServerSpec) {
	*out = *in
	if in.SSH != nil {
		in, out := &in.SSH, &out.SSH
		*out = new(GitServerSpecSSH)
		(*in).DeepCopyInto(*out)
	}
	if in.HTTP != nil {
		in, out := &in.HTTP, &out.HTTP
		*out = new(GitServerSpecHTTP)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpec.
func (in *GitServerSpec) DeepCopy() *GitServerSpec {
	if in == nil {
		return nil
	}
	out := new(GitServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpecHTTP) DeepCopyInto(out *GitServerSpecHTTP) {
	*out = *in
	in.Auth.DeepCopyInto(&out.Auth)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpecHTTP.
func (in *GitServerSpecHTTP) DeepCopy() *GitServerSpecHTTP {
	if in == nil {
		return nil
	}
	out := new(GitServerSpecHTTP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpecHTTPAuth) DeepCopyInto(out *GitServerSpecHTTPAuth) {
	*out = *in
	if in.Username != nil {
		in, out := &in.Username, &out.Username
		*out = new(GitServerSpecHTTPAuthUsername)
		**out = **in
	}
	if in.Password != nil {
		in, out := &in.Password, &out.Password
		*out = new(GitServerSpecHTTPAuthPassword)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpecHTTPAuth.
func (in *GitServerSpecHTTPAuth) DeepCopy() *GitServerSpecHTTPAuth {
	if in == nil {
		return nil
	}
	out := new(GitServerSpecHTTPAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpecHTTPAuthPassword) DeepCopyInto(out *GitServerSpecHTTPAuthPassword) {
	*out = *in
	out.ValueFrom = in.ValueFrom
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpecHTTPAuthPassword.
func (in *GitServerSpecHTTPAuthPassword) DeepCopy() *GitServerSpecHTTPAuthPassword {
	if in == nil {
		return nil
	}
	out := new(GitServerSpecHTTPAuthPassword)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpecHTTPAuthUsername) DeepCopyInto(out *GitServerSpecHTTPAuthUsername) {
	*out = *in
	out.ValueFrom = in.ValueFrom
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpecHTTPAuthUsername.
func (in *GitServerSpecHTTPAuthUsername) DeepCopy() *GitServerSpecHTTPAuthUsername {
	if in == nil {
		return nil
	}
	out := new(GitServerSpecHTTPAuthUsername)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpecSSH) DeepCopyInto(out *GitServerSpecSSH) {
	*out = *in
	in.Auth.DeepCopyInto(&out.Auth)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpecSSH.
func (in *GitServerSpecSSH) DeepCopy() *GitServerSpecSSH {
	if in == nil {
		return nil
	}
	out := new(GitServerSpecSSH)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpecSSHAuth) DeepCopyInto(out *GitServerSpecSSHAuth) {
	*out = *in
	if in.AuthorizedKeys != nil {
		in, out := &in.AuthorizedKeys, &out.AuthorizedKeys
		*out = new(GitServerSpecSSHAuthAuthorizedKeys)
		**out = **in
	}
	if in.HostKey != nil {
		in, out := &in.HostKey, &out.HostKey
		*out = new(GitServerSpecSSHAuthHostKey)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpecSSHAuth.
func (in *GitServerSpecSSHAuth) DeepCopy() *GitServerSpecSSHAuth {
	if in == nil {
		return nil
	}
	out := new(GitServerSpecSSHAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpecSSHAuthAuthorizedKeys) DeepCopyInto(out *GitServerSpecSSHAuthAuthorizedKeys) {
	*out = *in
	out.ValueFrom = in.ValueFrom
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpecSSHAuthAuthorizedKeys.
func (in *GitServerSpecSSHAuthAuthorizedKeys) DeepCopy() *GitServerSpecSSHAuthAuthorizedKeys {
	if in == nil {
		return nil
	}
	out := new(GitServerSpecSSHAuthAuthorizedKeys)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpecSSHAuthHostKey) DeepCopyInto(out *GitServerSpecSSHAuthHostKey) {
	*out = *in
	out.ValueFrom = in.ValueFrom
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpecSSHAuthHostKey.
func (in *GitServerSpecSSHAuthHostKey) DeepCopy() *GitServerSpecSSHAuthHostKey {
	if in == nil {
		return nil
	}
	out := new(GitServerSpecSSHAuthHostKey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerStatus) DeepCopyInto(out *GitServerStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	if in.DeploymentRef != nil {
		in, out := &in.DeploymentRef, &out.DeploymentRef
		*out = new(TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
	if in.ServiceRef != nil {
		in, out := &in.ServiceRef, &out.ServiceRef
		*out = new(TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
	if in.Address != nil {
		in, out := &in.Address, &out.Address
		*out = new(Addressable)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerStatus.
func (in *GitServerStatus) DeepCopy() *GitServerStatus {
	if in == nil {
		return nil
	}
	out := new(GitServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretKeyRef) DeepCopyInto(out *SecretKeyRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretKeyRef.
func (in *SecretKeyRef) DeepCopy() *SecretKeyRef {
	if in == nil {
		return nil
	}
	out := new(SecretKeyRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TypedLocalObjectReference) DeepCopyInto(out *TypedLocalObjectReference) {
	*out = *in
	if in.APIGroup != nil {
		in, out := &in.APIGroup, &out.APIGroup
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TypedLocalObjectReference.
func (in *TypedLocalObjectReference) DeepCopy() *TypedLocalObjectReference {
	if in == nil {
		return nil
	}
	out := new(TypedLocalObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueFrom) DeepCopyInto(out *ValueFrom) {
	*out = *in
	out.SecretKeyRef = in.SecretKeyRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueFrom.
func (in *ValueFrom) DeepCopy() *ValueFrom {
	if in == nil {
		return nil
	}
	out := new(ValueFrom)
	in.DeepCopyInto(out)
	return out
}
