package kptfile

// SetLabels sets the labels in the Kptfile
func (kf *Kptfile) SetLabels(labels map[string]string) {
	for k, v := range labels {
		_ = kf.SetLabel(k, v)
	}

	existing := kf.GetLabels()
	for k := range existing {
		if _, ok := labels[k]; !ok {
			_ = kf.RemoveLabel(k)
		}
	}
}

// SetAnnotations sets the annotations in the Kptfile
func (kf *Kptfile) SetAnnotations(annotations map[string]string) {
	for k, v := range annotations {
		_ = kf.SetAnnotation(k, v)
	}

	existing := kf.GetAnnotations()
	for k := range existing {
		if _, ok := annotations[k]; !ok {
			_ = kf.RemoveAnnotation(k)
		}
	}
}
