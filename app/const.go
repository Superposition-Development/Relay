package app

var (
	WebsocketEndpoint    = "/ws"
	SignupEndpoint       = "/signup"
	LoginEndpoint        = "/login"
	ValidateUserEndpoint = "/validateUserToken"

	CreateServerEndpoint   = "/createServer"
	GetServerEndpoint      = "/getServers"
	JoinServerEndpoint     = "/joinServer"
	GetServerUsersEndpoint = "/getServerUsers"

	CreateChannelEndpoint = "/createChannel"
	GetChannelEndpoint    = "/getChannels"

	CreateDMEndpoint = "/createDM"
	GetDMEndpoint    = "/getDMs"

	//socket endpoints
	SendMessageEndpoint       = "/sendMessage"
	GetMessagesServerEndpoint = "/getMessagesServer"
	GetMessagesDMEndpoint     = "/getMessagesDM"
)
