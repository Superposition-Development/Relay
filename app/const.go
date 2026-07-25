package app

var (
	WebsocketEndpoint    = "/ws"
	SignupEndpoint       = "/signup"
	LoginEndpoint        = "/login"
	ValidateUserEndpoint = "/validateUserToken"

	CreateServerEndpoint = "/createServer"
	GetServerEndpoint    = "/getServers"
	JoinServerEndpoint   = "/joinServer"

	CreateChannelEndpoint = "/createChannel"
	GetChannelEndpoint    = "/getChannels"

	//socket endpoints
	SendMessageEndpoint = "/sendMessage"
	GetMessagesEndpoint = "/getMessages"
)
