package main

import (
	"context"
	"fmt"

	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/container/processors"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
)

// ---------------------------------------------------------------------------
// 1. Interface definitions
// ---------------------------------------------------------------------------

type UserService interface {
	GetUser(id int) string
}

type OrderService interface {
	CreateOrder(userId int, item string) string
}

type NotifyService interface {
	Send(to, msg string)
}

// ---------------------------------------------------------------------------
// 2. Basic components with wire injection (pointer + interface)
// ---------------------------------------------------------------------------

type UserServiceImpl struct {
	Logger syslog.Logger `logger:""`
	DB     *Database     `wire:""`
}

func (s *UserServiceImpl) GetUser(id int) string {
	s.Logger.Infof("GetUser(%d) via %s", id, s.DB.DSN)
	return fmt.Sprintf("user-%d", id)
}

func (s *UserServiceImpl) AfterPropertiesSet() error {
	s.Logger.Info("UserServiceImpl.AfterPropertiesSet called")
	return nil
}

// ---------------------------------------------------------------------------
// 3. Component with interface injection + slice injection
// ---------------------------------------------------------------------------

type OrderServiceImpl struct {
	Logger    syslog.Logger   `logger:""`
	Users     UserService     `wire:""`           // interface injection
	Notifiers []NotifyService `wire:""`           // slice interface injection
	MaxRetry  int             `value:"${order.max_retry:3}"`
}

func (s *OrderServiceImpl) CreateOrder(userId int, item string) string {
	user := s.Users.GetUser(userId)
	orderId := fmt.Sprintf("ORD-%s-%s", user, item)
	for _, n := range s.Notifiers {
		n.Send(user, fmt.Sprintf("Order %s created (maxRetry=%d)", orderId, s.MaxRetry))
	}
	return orderId
}

func (s *OrderServiceImpl) Init() error {
	s.Logger.Infof("OrderServiceImpl.Init: maxRetry=%d, notifiers=%d", s.MaxRetry, len(s.Notifiers))
	return nil
}

// ---------------------------------------------------------------------------
// 4. Multiple implementations of NotifyService (Primary + Qualifier)
// ---------------------------------------------------------------------------

type EmailNotifier struct {
	Logger syslog.Logger `logger:""`
	SMTP   string        `value:"${notify.email.smtp:smtp.example.com}"`
}

func (n *EmailNotifier) Send(to, msg string) {
	n.Logger.Infof("[Email via %s] to=%s msg=%s", n.SMTP, to, msg)
}

func (n *EmailNotifier) Primary() {} // preferred when single injection

func (n *EmailNotifier) Qualifier() string { return "email" }

type SMSNotifier struct {
	Logger   syslog.Logger `logger:""`
	Gateway  string        `value:"${notify.sms.gateway:sms.example.com}"`
}

func (n *SMSNotifier) Send(to, msg string) {
	n.Logger.Infof("[SMS via %s] to=%s msg=%s", n.Gateway, to, msg)
}

func (n *SMSNotifier) Qualifier() string { return "sms" }

// ---------------------------------------------------------------------------
// 5. Database component with configuration properties (prefix tag)
// ---------------------------------------------------------------------------

type Database struct {
	DSN     string `yaml:"dsn"`
	MaxConn int    `yaml:"max_conn"`
}

func (d *Database) Prefix() string { return "database" }

func (d *Database) AfterPropertiesSet() error {
	fmt.Printf("[Database] connected: dsn=%s maxConn=%d\n", d.DSN, d.MaxConn)
	return nil
}

// ---------------------------------------------------------------------------
// 6. Constructor injection
// ---------------------------------------------------------------------------

type Analytics struct {
	orders OrderService
	users  UserService
}

func NewAnalytics(o OrderService, u UserService) *Analytics {
	return &Analytics{orders: o, users: u}
}

func (a *Analytics) Report(userId int) string {
	user := a.users.GetUser(userId)
	return fmt.Sprintf("report for %s", user)
}

func (a *Analytics) Init() error {
	fmt.Println("[Analytics] initialized via constructor injection")
	return nil
}

// ---------------------------------------------------------------------------
// 7. Embedded struct injection
// ---------------------------------------------------------------------------

type BaseComponent struct {
	Logger syslog.Logger `logger:",embed"`
}

type CacheService struct {
	BaseComponent
	DB *Database `wire:""`
}

func (c *CacheService) Init() error {
	c.Logger.Infof("CacheService.Init: caching for dsn=%s", c.DB.DSN)
	return nil
}

// ---------------------------------------------------------------------------
// 8. Named component with custom naming
// ---------------------------------------------------------------------------

type HealthChecker struct {
	Logger syslog.Logger `logger:""`
	DB     *Database     `wire:""`
}

func (h *HealthChecker) Naming() string { return "health-checker" }

func (h *HealthChecker) Init() error {
	h.Logger.Info("HealthChecker ready")
	return nil
}

// ---------------------------------------------------------------------------
// 9. Optional injection (required=false)
// ---------------------------------------------------------------------------

type MetricsCollector struct {
	Logger syslog.Logger `logger:""`
	Tracer *Tracer       `wire:",required=false"` // won't fail if missing
}

func (m *MetricsCollector) Init() error {
	if m.Tracer != nil {
		m.Logger.Info("MetricsCollector: tracer attached")
	} else {
		m.Logger.Info("MetricsCollector: running without tracer")
	}
	return nil
}

type Tracer struct{} // not registered — tests required=false

// ---------------------------------------------------------------------------
// 10. ApplicationRunner (ordered startup)
// ---------------------------------------------------------------------------

type AppBootstrap struct {
	Logger syslog.Logger `logger:""`
	Orders OrderService  `wire:""`
}

func (r *AppBootstrap) Run() error {
	r.Logger.Info("AppBootstrap.Run: application is ready")
	orderId := r.Orders.CreateOrder(1, "book")
	r.Logger.Infof("sample order created: %s", orderId)
	return nil
}

func (r *AppBootstrap) Order() int { return 1 }

// ---------------------------------------------------------------------------
// 11. Custom PostProcessor — adds logging around Init
// ---------------------------------------------------------------------------

type initLoggingProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
}

func (p *initLoggingProcessor) PostProcessBeforeInitialization(component any, componentName string) (any, error) {
	if _, ok := component.(definition.InitializeComponent); ok {
		fmt.Printf("  [PostProcessor] >> before Init() of '%s'\n", componentName)
	}
	return component, nil
}

func (p *initLoggingProcessor) PostProcessAfterInitialization(component any, componentName string) (any, error) {
	if _, ok := component.(definition.InitializeComponent); ok {
		fmt.Printf("  [PostProcessor] << after Init() of '%s'\n", componentName)
	}
	return component, nil
}

// ---------------------------------------------------------------------------
// 12. ConditionalComponent — only registers when config says so
// ---------------------------------------------------------------------------

type FeatureFlag struct {
	Logger syslog.Logger `logger:""`
}

func (f *FeatureFlag) Condition(ctx definition.ConditionContext) bool {
	val := ctx.GetConfig("feature.beta_enabled")
	return val == true
}

func (f *FeatureFlag) Init() error {
	f.Logger.Info("FeatureFlag: beta features enabled")
	return nil
}

// ---------------------------------------------------------------------------
// 13. Event listener
// ---------------------------------------------------------------------------

type startupListener struct{}

func (l *startupListener) OnEvent(event definition.ApplicationEvent) error {
	switch event.(type) {
	case *definition.ApplicationStartedEvent:
		fmt.Println("[Event] ApplicationStarted received")
	case *definition.ApplicationClosingEvent:
		fmt.Println("[Event] ApplicationClosing received")
	}
	return nil
}

// ---------------------------------------------------------------------------
// 14. CloserComponent
// ---------------------------------------------------------------------------

type ResourceCleaner struct {
	Logger syslog.Logger `logger:""`
}

func (c *ResourceCleaner) CloseWithContext(_ context.Context) error {
	c.Logger.Info("ResourceCleaner: cleaning up resources")
	return nil
}

// ---------------------------------------------------------------------------
// main — run with: go run . --ioc:run_debug
// ---------------------------------------------------------------------------

func main() {
	config := []byte(`
database:
  dsn: "postgres://localhost:5432/demo"
  max_conn: 10
order:
  max_retry: 5
notify:
  email:
    smtp: "smtp.mymail.com"
  sms:
    gateway: "sms.provider.io"
feature:
  beta_enabled: true
`)

	a, err := ioc.RunDebug(
		app.LogTrace,
		app.AddConfigLoader(loader.NewRawLoader(config)),
		app.SetComponents(
			// database
			&Database{},
			// services
			&UserServiceImpl{},
			&OrderServiceImpl{},
			// notifiers (multiple implementations)
			&EmailNotifier{},
			&SMSNotifier{},
			// constructor injection
			NewAnalytics,
			// embedded struct
			&CacheService{},
			// named component
			&HealthChecker{},
			// optional injection
			&MetricsCollector{},
			// runner
			&AppBootstrap{},
			// custom post processor
			&initLoggingProcessor{},
			// conditional component
			&FeatureFlag{},
			// event listener
			&startupListener{},
			// closer
			&ResourceCleaner{},
		),
	)
	if err != nil {
		panic(err)
	}
	defer a.Close()

	fmt.Println("\n=== Application is running. Press Ctrl+C to exit ===")
	select {}
}
