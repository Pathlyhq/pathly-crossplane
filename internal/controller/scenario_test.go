package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"

	"github.com/pathlyhq/pathly-crossplane/apis/pathly/v1alpha1"
	pathlyclient "github.com/pathlyhq/pathly-crossplane/internal/client"
)

func schemeWithPathly(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return s
}

func testManager(t *testing.T) ctrl.Manager {
	t.Helper()
	mgr, err := ctrl.NewManager(&rest.Config{}, ctrl.Options{
		Scheme:                 schemeWithPathly(t),
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return mgr
}

func TestSetupAndDefaultOptions(t *testing.T) {
	mgr := testManager(t)
	o := DefaultOptions(logging.NewNopLogger())
	if err := Setup(mgr, o); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := setupAll(mgr, o, func(ctrl.Manager, controller.Options) error {
		return errors.New("boom")
	}); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("setupAll must surface setup errors: %v", err)
	}
}

func TestConnectorsConnectAndDisconnect(t *testing.T) {
	s := schemeWithPathly(t)
	base := "https://api.example.com/"
	pc := &v1alpha1.ProviderConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
		Spec: v1alpha1.ProviderConfigSpec{
			CredentialsSource: xpv1.CredentialsSourceSecret,
			BaseURL:           &base,
			CredentialsSecretRef: &xpv1.SecretKeySelector{
				SecretReference: xpv1.SecretReference{Namespace: "ns", Name: "creds"},
				Key:             "token",
			},
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "creds"},
		Data:       map[string][]byte{"token": []byte("sp_x")},
	}
	kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc, secret).Build()

	mg := &v1alpha1.Scenario{
		ObjectMeta: metav1.ObjectMeta{Name: "s"},
		Spec: v1alpha1.ScenarioSpec{
			ResourceSpec: xpv1.ResourceSpec{ProviderConfigReference: &xpv1.Reference{Name: "default"}},
		},
	}

	for _, c := range []managed.ExternalConnecter{
		&scenarioConnector{kube: kube},
		&webhookConnector{kube: kube},
		&mwConnector{kube: kube},
		&slaConnector{kube: kube},
	} {
		ext, err := c.Connect(context.Background(), mg)
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}
		if err := ext.Disconnect(context.Background()); err != nil {
			t.Fatalf("Disconnect: %v", err)
		}
	}

	// Connect fails when ProviderConfig is missing.
	empty := fake.NewClientBuilder().WithScheme(s).Build()
	for _, c := range []managed.ExternalConnecter{
		&scenarioConnector{kube: empty},
		&webhookConnector{kube: empty},
		&mwConnector{kube: empty},
		&slaConnector{kube: empty},
	} {
		if _, err := c.Connect(context.Background(), mg); err == nil {
			t.Fatal("Connect must fail without ProviderConfig")
		}
	}
}

func TestNewPathlyClientCredentialPaths(t *testing.T) {
	s := schemeWithPathly(t)
	ctx := context.Background()

	mgWith := func(pcName string) *v1alpha1.Scenario {
		return &v1alpha1.Scenario{
			ObjectMeta: metav1.ObjectMeta{Name: "s"},
			Spec: v1alpha1.ScenarioSpec{
				ResourceSpec: xpv1.ResourceSpec{ProviderConfigReference: &xpv1.Reference{Name: pcName}},
			},
		}
	}

	t.Run("missing provider config", func(t *testing.T) {
		kube := fake.NewClientBuilder().WithScheme(s).Build()
		_, err := NewPathlyClient(ctx, kube, mgWith("missing"))
		if err == nil || !strings.Contains(err.Error(), errGetProviderConfig) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("unsupported credentials source", func(t *testing.T) {
		pc := &v1alpha1.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "pc"},
			Spec:       v1alpha1.ProviderConfigSpec{CredentialsSource: xpv1.CredentialsSourceNone},
		}
		kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc).Build()
		_, err := NewPathlyClient(ctx, kube, mgWith("pc"))
		if err == nil || !strings.Contains(err.Error(), "not supported") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("missing secret ref", func(t *testing.T) {
		pc := &v1alpha1.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "pc"},
			Spec:       v1alpha1.ProviderConfigSpec{CredentialsSource: xpv1.CredentialsSourceSecret},
		}
		kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc).Build()
		_, err := NewPathlyClient(ctx, kube, mgWith("pc"))
		if err == nil || !strings.Contains(err.Error(), "credentialsSecretRef") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("missing secret object", func(t *testing.T) {
		pc := &v1alpha1.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "pc"},
			Spec: v1alpha1.ProviderConfigSpec{
				CredentialsSource: xpv1.CredentialsSourceSecret,
				CredentialsSecretRef: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{Namespace: "ns", Name: "gone"},
					Key:             "token",
				},
			},
		}
		kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc).Build()
		_, err := NewPathlyClient(ctx, kube, mgWith("pc"))
		if err == nil || !strings.Contains(err.Error(), errGetSecret) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("empty secret defaults to token then apiToken", func(t *testing.T) {
		pc := &v1alpha1.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "pc"},
			Spec: v1alpha1.ProviderConfigSpec{
				// empty credentialsSource defaults to Secret
				CredentialsSecretRef: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{Namespace: "ns", Name: "creds"},
					// empty key defaults to "token"
				},
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "creds"},
			Data:       map[string][]byte{"apiToken": []byte(" sp_from_api ")},
		}
		kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc, secret).Build()
		cli, err := NewPathlyClient(ctx, kube, mgWith("pc"))
		if err != nil {
			t.Fatalf("NewPathlyClient: %v", err)
		}
		if cli == nil {
			t.Fatal("expected client")
		}
	})

	t.Run("no token keys", func(t *testing.T) {
		pc := &v1alpha1.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "pc"},
			Spec: v1alpha1.ProviderConfigSpec{
				CredentialsSource: xpv1.CredentialsSourceSecret,
				CredentialsSecretRef: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{Namespace: "ns", Name: "creds"},
					Key:             "token",
				},
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "creds"},
			Data:       map[string][]byte{"other": []byte("x")},
		}
		kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc, secret).Build()
		_, err := NewPathlyClient(ctx, kube, mgWith("pc"))
		if err == nil || !strings.Contains(err.Error(), errNoToken) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("apiToken key preferred when selected", func(t *testing.T) {
		pc := &v1alpha1.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "pc"},
			Spec: v1alpha1.ProviderConfigSpec{
				CredentialsSource: xpv1.CredentialsSourceSecret,
				CredentialsSecretRef: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{Namespace: "ns", Name: "creds"},
					Key:             "apiToken",
				},
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "creds"},
			Data:       map[string][]byte{"apiToken": []byte("sp_direct")},
		}
		kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc, secret).Build()
		if _, err := NewPathlyClient(ctx, kube, mgWith("pc")); err != nil {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("apiToken key missing yields no token", func(t *testing.T) {
		pc := &v1alpha1.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "pc"},
			Spec: v1alpha1.ProviderConfigSpec{
				CredentialsSource: xpv1.CredentialsSourceSecret,
				CredentialsSecretRef: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{Namespace: "ns", Name: "creds"},
					Key:             "apiToken",
				},
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "creds"},
			Data:       map[string][]byte{"token": []byte("sp_other")},
		}
		kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc, secret).Build()
		_, err := NewPathlyClient(ctx, kube, mgWith("pc"))
		if err == nil || !strings.Contains(err.Error(), errNoToken) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("blank baseURL ignored", func(t *testing.T) {
		blank := "   "
		pc := &v1alpha1.ProviderConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "pc"},
			Spec: v1alpha1.ProviderConfigSpec{
				CredentialsSource: xpv1.CredentialsSourceSecret,
				BaseURL:           &blank,
				CredentialsSecretRef: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{Namespace: "ns", Name: "creds"},
					Key:             "token",
				},
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "creds"},
			Data:       map[string][]byte{"token": []byte("sp_x")},
		}
		kube := fake.NewClientBuilder().WithScheme(s).WithObjects(pc, secret).Build()
		if _, err := NewPathlyClient(ctx, kube, mgWith("pc")); err != nil {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("missing providerConfigRef", func(t *testing.T) {
		kube := fake.NewClientBuilder().WithScheme(s).Build()
		_, err := NewPathlyClient(ctx, kube, &v1alpha1.Scenario{ObjectMeta: metav1.ObjectMeta{Name: "x"}})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestScenarioExternalCRUD(t *testing.T) {
	var createdBody map[string]any
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/scenarios":
			gotKey = r.Header.Get("Idempotency-Key")
			_ = json.NewDecoder(r.Body).Decode(&createdBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"mon-1","name":"Checkout","type":"http","url":"https://example.com","enabled":true,"method":"GET","expectedStatus":200,"intervalSec":300}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/scenarios/mon-1":
			_, _ = w.Write([]byte(`{"id":"mon-1","name":"Checkout","type":"http","url":"https://example.com","enabled":true,"method":"GET","expectedStatus":200,"intervalSec":300}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/v1/scenarios/mon-1":
			_, _ = w.Write([]byte(`{"id":"mon-1","name":"Checkout","type":"http"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/scenarios/mon-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer srv.Close()

	ext := &scenarioExternal{client: pathlyclient.New(srv.URL, "sp_x")}
	enabled := true
	method := "GET"
	status := int64(200)
	interval := int64(300)
	cr := &v1alpha1.Scenario{
		ObjectMeta: metav1.ObjectMeta{Name: "checkout", UID: "uid-1"},
		Spec: v1alpha1.ScenarioSpec{
			ForProvider: v1alpha1.ScenarioParameters{
				Name: "Checkout", Type: "http",
				URL: StringPtr("https://example.com"), Enabled: &enabled,
				Method: &method, ExpectedStatus: &status, IntervalSec: &interval,
				Regions: []string{"europe-west1"}, Tags: []string{"prod"},
			},
		},
	}

	obs, err := ext.Observe(context.Background(), cr)
	if err != nil || obs.ResourceExists {
		t.Fatalf("Observe empty: %+v err=%v", obs, err)
	}
	if _, err := ext.Create(context.Background(), cr); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if meta.GetExternalName(cr) != "mon-1" || gotKey == "" || createdBody["name"] != "Checkout" {
		t.Fatalf("create state: ext=%q key=%q body=%v", meta.GetExternalName(cr), gotKey, createdBody)
	}
	obs, err = ext.Observe(context.Background(), cr)
	if err != nil || !obs.ResourceExists || !obs.ResourceUpToDate {
		t.Fatalf("Observe: %+v err=%v", obs, err)
	}
	if _, err := ext.Update(context.Background(), cr); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := ext.Delete(context.Background(), cr); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Delete with empty external name is a no-op success.
	empty := &v1alpha1.Scenario{ObjectMeta: metav1.ObjectMeta{Name: "e"}}
	if _, err := ext.Delete(context.Background(), empty); err != nil {
		t.Fatalf("Delete empty: %v", err)
	}
}

func TestScenarioAPIErrorsAndDrift(t *testing.T) {
	forbid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"missing scope"}`))
	}))
	defer forbid.Close()
	ext := &scenarioExternal{client: pathlyclient.New(forbid.URL, "sp_x")}
	cr := &v1alpha1.Scenario{
		ObjectMeta: metav1.ObjectMeta{Name: "x", UID: "u"},
		Spec:       v1alpha1.ScenarioSpec{ForProvider: v1alpha1.ScenarioParameters{Name: "X", Type: "http"}},
	}
	meta.SetExternalName(cr, "mon-1")
	if _, err := ext.Observe(context.Background(), cr); err == nil {
		t.Fatal("Observe must surface API errors")
	}
	if _, err := ext.Create(context.Background(), cr); err == nil {
		t.Fatal("Create must surface API errors")
	}
	if _, err := ext.Update(context.Background(), cr); err == nil {
		t.Fatal("Update must surface API errors")
	}
	if _, err := ext.Delete(context.Background(), cr); err == nil {
		t.Fatal("Delete must surface API errors")
	}

	// Not-found on delete is ignored.
	nf := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"gone"}`))
	}))
	defer nf.Close()
	extNF := &scenarioExternal{client: pathlyclient.New(nf.URL, "sp_x")}
	meta.SetExternalName(cr, "mon-gone")
	if _, err := extNF.Delete(context.Background(), cr); err != nil {
		t.Fatalf("Delete not-found: %v", err)
	}
	obs, err := extNF.Observe(context.Background(), cr)
	if err != nil || obs.ResourceExists {
		t.Fatalf("Observe not-found: %+v err=%v", obs, err)
	}

	wrong := &v1alpha1.Webhook{}
	for _, call := range []func() error{
		func() error { _, err := ext.Observe(context.Background(), wrong); return err },
		func() error { _, err := ext.Create(context.Background(), wrong); return err },
		func() error { _, err := ext.Update(context.Background(), wrong); return err },
		func() error { _, err := ext.Delete(context.Background(), wrong); return err },
	} {
		if err := call(); err == nil {
			t.Fatal("wrong type must error")
		}
	}
}

func TestScenarioUpToDateAndInputDefaults(t *testing.T) {
	url := "https://a.example"
	enabled := true
	interval := int64(60)
	method := "POST"
	status := int64(201)
	want := v1alpha1.ScenarioParameters{
		Name: "A", URL: &url, Enabled: &enabled, IntervalSec: &interval, Method: &method, ExpectedStatus: &status,
	}
	got := &pathlyclient.Scenario{
		Name: "A", URL: &url, Enabled: &enabled, IntervalSec: &interval, Method: &method, ExpectedStatus: &status,
	}
	if !scenarioUpToDate(want, got) {
		t.Fatal("expected up to date")
	}
	got.Name = "B"
	if scenarioUpToDate(want, got) {
		t.Fatal("name drift")
	}
	got.Name = "A"
	other := "https://b.example"
	got.URL = &other
	if scenarioUpToDate(want, got) {
		t.Fatal("url drift")
	}
	got.URL = nil
	if scenarioUpToDate(want, got) {
		t.Fatal("nil url drift")
	}
	got.URL = &url
	off := false
	got.Enabled = &off
	if scenarioUpToDate(want, got) {
		t.Fatal("enabled drift")
	}
	got.Enabled = nil
	if scenarioUpToDate(want, got) {
		t.Fatal("nil enabled drift")
	}
	got.Enabled = &enabled
	got.IntervalSec = nil
	if scenarioUpToDate(want, got) {
		t.Fatal("nil interval drift")
	}
	wrongI := int64(1)
	got.IntervalSec = &wrongI
	if scenarioUpToDate(want, got) {
		t.Fatal("interval drift")
	}
	got.IntervalSec = &interval
	got.Method = nil
	if scenarioUpToDate(want, got) {
		t.Fatal("nil method drift")
	}
	m2 := "GET"
	got.Method = &m2
	if scenarioUpToDate(want, got) {
		t.Fatal("method drift")
	}
	got.Method = &method
	got.ExpectedStatus = nil
	if scenarioUpToDate(want, got) {
		t.Fatal("nil status drift")
	}
	st2 := int64(500)
	got.ExpectedStatus = &st2
	if scenarioUpToDate(want, got) {
		t.Fatal("status drift")
	}

	in := scenarioInput(v1alpha1.ScenarioParameters{Name: "N"}) // type defaults to http
	if in.Type == nil || *in.Type != "http" {
		t.Fatalf("default type = %v", in.Type)
	}
}

func TestWebhookExternalFull(t *testing.T) {
	var deleted bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"h1","events":["run.failed"],"secret":"whsec_x","urlFingerprint":"fp","enabled":true,"hasSecret":true,"createdAt":"2026-01-01T00:00:00Z"}`))
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"id":"h1","events":["run.failed"],"hasSecret":true,"urlFingerprint":"fp","enabled":true,"createdAt":"2026-01-01T00:00:00Z"}`))
		case http.MethodDelete:
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	ext := &webhookExternal{client: pathlyclient.New(srv.URL, "sp_x")}
	cr := &v1alpha1.Webhook{
		ObjectMeta: metav1.ObjectMeta{Name: "hook", UID: "uid-w"},
		Spec: v1alpha1.WebhookSpec{
			ForProvider: v1alpha1.WebhookParameters{URL: "https://hooks.example.com", Events: []string{"run.failed"}},
		},
	}

	obs, err := ext.Observe(context.Background(), cr)
	if err != nil || obs.ResourceExists {
		t.Fatalf("Observe empty: %+v err=%v", obs, err)
	}
	creation, err := ext.Create(context.Background(), cr)
	if err != nil || string(creation.ConnectionDetails["secret"]) != "whsec_x" {
		t.Fatalf("Create: %v details=%v", err, creation.ConnectionDetails)
	}
	obs, err = ext.Observe(context.Background(), cr)
	if err != nil || !obs.ResourceExists || !obs.ResourceUpToDate {
		t.Fatalf("Observe: %+v err=%v", obs, err)
	}
	if _, err := ext.Update(context.Background(), cr); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := ext.Delete(context.Background(), cr); err != nil || !deleted {
		t.Fatalf("Delete: err=%v deleted=%v", err, deleted)
	}

	// Recreate after delete: create again without secret in response still ok.
	nosec := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"h2","events":["run.failed"]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"gone"}`))
	}))
	defer nosec.Close()
	ext2 := &webhookExternal{client: pathlyclient.New(nosec.URL, "sp_x")}
	cr2 := &v1alpha1.Webhook{
		ObjectMeta: metav1.ObjectMeta{Name: "hook2", UID: "uid-w2"},
		Spec:       v1alpha1.WebhookSpec{ForProvider: v1alpha1.WebhookParameters{URL: "https://h.example"}},
	}
	creation, err = ext2.Create(context.Background(), cr2)
	if err != nil {
		t.Fatalf("recreate: %v", err)
	}
	if _, ok := creation.ConnectionDetails["secret"]; ok {
		t.Fatal("no secret expected")
	}
	meta.SetExternalName(cr2, "h-gone")
	obs, err = ext2.Observe(context.Background(), cr2)
	if err != nil || obs.ResourceExists {
		t.Fatalf("observe gone: %+v err=%v", obs, err)
	}
	if _, err := ext2.Delete(context.Background(), cr2); err != nil {
		t.Fatalf("delete not-found: %v", err)
	}
	if _, err := ext2.Delete(context.Background(), &v1alpha1.Webhook{ObjectMeta: metav1.ObjectMeta{Name: "e"}}); err != nil {
		t.Fatalf("delete empty: %v", err)
	}

	forbid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"no"}`))
	}))
	defer forbid.Close()
	bad := &webhookExternal{client: pathlyclient.New(forbid.URL, "sp_x")}
	wh := &v1alpha1.Webhook{ObjectMeta: metav1.ObjectMeta{Name: "w", UID: "u"}, Spec: v1alpha1.WebhookSpec{ForProvider: v1alpha1.WebhookParameters{URL: "https://x"}}}
	meta.SetExternalName(wh, "h1")
	if _, err := bad.Observe(context.Background(), wh); err == nil {
		t.Fatal("observe api error")
	}
	if _, err := bad.Create(context.Background(), wh); err == nil {
		t.Fatal("create api error")
	}
	if _, err := bad.Delete(context.Background(), wh); err == nil {
		t.Fatal("delete api error")
	}
	wrong := &v1alpha1.Scenario{}
	if _, err := bad.Observe(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
	if _, err := bad.Create(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
	if _, err := bad.Delete(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
}

func TestMaintenanceWindowExternalFull(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_, _ = w.Write([]byte(`{"id":"mw-1","monitorId":"mon-1","startsAt":"2026-01-01T00:00:00Z","endsAt":"2026-01-01T01:00:00Z","reason":"r","weekday":6,"startMinute":120,"durationMin":90}`))
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"id":"mw-1","monitorId":"mon-1","startsAt":"2026-01-01T00:00:00Z","endsAt":"2026-01-01T01:00:00Z","reason":"r","weekday":6,"startMinute":120,"durationMin":90}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	ext := &mwExternal{client: pathlyclient.New(srv.URL, "sp_x")}
	weekday, start, dur := int64(6), int64(120), int64(90)
	mon := "mon-1"
	cr := &v1alpha1.MaintenanceWindow{
		ObjectMeta: metav1.ObjectMeta{Name: "mw", UID: "uid-m"},
		Spec: v1alpha1.MaintenanceWindowSpec{ForProvider: v1alpha1.MaintenanceWindowParameters{
			MonitorID: &mon, StartsAt: StringPtr("2026-01-01T00:00:00Z"), EndsAt: StringPtr("2026-01-01T01:00:00Z"),
			Reason: StringPtr("r"), Weekday: &weekday, StartMinute: &start, DurationMin: &dur,
		}},
	}
	obs, err := ext.Observe(context.Background(), cr)
	if err != nil || obs.ResourceExists {
		t.Fatalf("empty: %+v err=%v", obs, err)
	}
	if _, err := ext.Create(context.Background(), cr); err != nil {
		t.Fatalf("Create: %v", err)
	}
	obs, err = ext.Observe(context.Background(), cr)
	if err != nil || !obs.ResourceExists {
		t.Fatalf("Observe: %+v err=%v", obs, err)
	}
	if _, err := ext.Update(context.Background(), cr); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := ext.Delete(context.Background(), cr); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Recreate path + not-found observe/delete + empty delete + API errors + wrong type.
	nf := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"id":"mw-2"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"gone"}`))
	}))
	defer nf.Close()
	ext2 := &mwExternal{client: pathlyclient.New(nf.URL, "sp_x")}
	cr2 := &v1alpha1.MaintenanceWindow{ObjectMeta: metav1.ObjectMeta{Name: "mw2", UID: "u2"}}
	if _, err := ext2.Create(context.Background(), cr2); err != nil {
		t.Fatalf("recreate: %v", err)
	}
	meta.SetExternalName(cr2, "mw-gone")
	obs, err = ext2.Observe(context.Background(), cr2)
	if err != nil || obs.ResourceExists {
		t.Fatalf("nf observe: %+v err=%v", obs, err)
	}
	if _, err := ext2.Delete(context.Background(), cr2); err != nil {
		t.Fatalf("nf delete: %v", err)
	}
	if _, err := ext2.Delete(context.Background(), &v1alpha1.MaintenanceWindow{ObjectMeta: metav1.ObjectMeta{Name: "e"}}); err != nil {
		t.Fatalf("empty delete: %v", err)
	}

	forbid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"no"}`))
	}))
	defer forbid.Close()
	bad := &mwExternal{client: pathlyclient.New(forbid.URL, "sp_x")}
	mw := &v1alpha1.MaintenanceWindow{ObjectMeta: metav1.ObjectMeta{Name: "m", UID: "u"}}
	meta.SetExternalName(mw, "mw-1")
	if _, err := bad.Observe(context.Background(), mw); err == nil {
		t.Fatal("observe err")
	}
	if _, err := bad.Create(context.Background(), mw); err == nil {
		t.Fatal("create err")
	}
	if _, err := bad.Delete(context.Background(), mw); err == nil {
		t.Fatal("delete err")
	}
	wrong := &v1alpha1.Scenario{}
	if _, err := bad.Observe(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
	if _, err := bad.Create(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
	if _, err := bad.Delete(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
}

func TestSlaTargetExternalFull(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			_, _ = w.Write([]byte(`{"id":"sla-1","monitorId":"mon-1","name":"Checkout","objectivePct":99.9,"windowDays":30,"excludeMaintenance":true,"warnAtBudgetRatio":0.8,"enabled":true}`))
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"id":"sla-1","monitorId":"mon-1","name":"Checkout","objectivePct":99.9,"windowDays":30,"excludeMaintenance":true,"warnAtBudgetRatio":0.8,"enabled":true}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	ext := &slaExternal{client: pathlyclient.New(srv.URL, "sp_x")}
	mon := "mon-1"
	name := "Checkout"
	ex, warn, en := true, 0.8, true
	cr := &v1alpha1.SlaTarget{
		ObjectMeta: metav1.ObjectMeta{Name: "sla", UID: "uid-s"},
		Spec: v1alpha1.SlaTargetSpec{ForProvider: v1alpha1.SlaTargetParameters{
			MonitorID: &mon, Name: &name, ObjectivePct: 99.9, WindowDays: 30,
			ExcludeMaintenance: &ex, WarnAtBudgetRatio: &warn, Enabled: &en,
		}},
	}
	obs, err := ext.Observe(context.Background(), cr)
	if err != nil || obs.ResourceExists {
		t.Fatalf("empty: %+v err=%v", obs, err)
	}
	if _, err := ext.Create(context.Background(), cr); err != nil {
		t.Fatalf("Create/upsert: %v", err)
	}
	obs, err = ext.Observe(context.Background(), cr)
	if err != nil || !obs.ResourceExists || !obs.ResourceUpToDate {
		t.Fatalf("Observe: %+v err=%v", obs, err)
	}
	cr.Spec.ForProvider.ObjectivePct = 99.5
	obs, err = ext.Observe(context.Background(), cr)
	if err != nil || obs.ResourceUpToDate {
		t.Fatalf("objective drift: %+v err=%v", obs, err)
	}
	cr.Spec.ForProvider.ObjectivePct = 99.9
	cr.Spec.ForProvider.WindowDays = 7
	obs, err = ext.Observe(context.Background(), cr)
	if err != nil || obs.ResourceUpToDate {
		t.Fatalf("window drift: %+v err=%v", obs, err)
	}
	cr.Spec.ForProvider.WindowDays = 30
	if _, err := ext.Update(context.Background(), cr); err != nil {
		t.Fatalf("Update upsert: %v", err)
	}
	if _, err := ext.Delete(context.Background(), cr); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	driftNil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"sla-1"}`))
	}))
	defer driftNil.Close()
	extNil := &slaExternal{client: pathlyclient.New(driftNil.URL, "sp_x")}
	meta.SetExternalName(cr, "sla-1")
	obs, err = extNil.Observe(context.Background(), cr)
	if err != nil || obs.ResourceUpToDate {
		t.Fatalf("nil fields drift: %+v err=%v", obs, err)
	}

	nf := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"gone"}`))
	}))
	defer nf.Close()
	extNF := &slaExternal{client: pathlyclient.New(nf.URL, "sp_x")}
	obs, err = extNF.Observe(context.Background(), cr)
	if err != nil || obs.ResourceExists {
		t.Fatalf("nf: %+v err=%v", obs, err)
	}
	if _, err := extNF.Delete(context.Background(), cr); err != nil {
		t.Fatalf("nf delete: %v", err)
	}
	if _, err := extNF.Delete(context.Background(), &v1alpha1.SlaTarget{ObjectMeta: metav1.ObjectMeta{Name: "e"}}); err != nil {
		t.Fatalf("empty delete: %v", err)
	}

	forbid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"no"}`))
	}))
	defer forbid.Close()
	bad := &slaExternal{client: pathlyclient.New(forbid.URL, "sp_x")}
	if _, err := bad.Observe(context.Background(), cr); err == nil {
		t.Fatal("observe")
	}
	if _, err := bad.Create(context.Background(), cr); err == nil {
		t.Fatal("create")
	}
	if _, err := bad.Update(context.Background(), cr); err == nil {
		t.Fatal("update")
	}
	if _, err := bad.Delete(context.Background(), cr); err == nil {
		t.Fatal("delete")
	}
	wrong := &v1alpha1.Scenario{}
	if _, err := bad.Observe(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
	if _, err := bad.Create(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
	if _, err := bad.Update(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
	if _, err := bad.Delete(context.Background(), wrong); err == nil {
		t.Fatal("type")
	}
}

func TestHelpers(t *testing.T) {
	if StringPtr("") != nil {
		t.Error("empty")
	}
	if StringPtr("a") == nil || *StringPtr("a") != "a" {
		t.Error("StringPtr")
	}
	s := "x"
	if diff := cmp.Diff("x", *CopyStringPtr(&s)); diff != "" {
		t.Error(diff)
	}
	i := int64(1)
	if *CopyInt64Ptr(&i) != 1 {
		t.Error("int")
	}
	b := true
	if !*CopyBoolPtr(&b) {
		t.Error("bool")
	}
	f := 1.5
	if *CopyFloat64Ptr(&f) != 1.5 {
		t.Error("float")
	}
	if CopyStringPtr(nil) != nil || CopyInt64Ptr(nil) != nil || CopyBoolPtr(nil) != nil || CopyFloat64Ptr(nil) != nil {
		t.Error("nil")
	}
}

var (
	_ managed.ExternalClient    = (*scenarioExternal)(nil)
	_ managed.ExternalClient    = (*webhookExternal)(nil)
	_ managed.ExternalClient    = (*mwExternal)(nil)
	_ managed.ExternalClient    = (*slaExternal)(nil)
	_ managed.ExternalConnecter = (*scenarioConnector)(nil)
	_ resource.Managed          = (*v1alpha1.Scenario)(nil)
)
