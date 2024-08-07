package auth

// BlacklistService implements the functionality of handling token blacklisting.
type BlacklistService struct {
	redisClient *BlacklistRedisClient
}

// NewBlacklistService creates a new BlacklistService value.
func NewBlacklistService(redisClient *BlacklistRedisClient) *BlacklistService {
	return &BlacklistService{redisClient: redisClient}
}

// BlacklistToken blacklists the given token.
func (s *BlacklistService) BlacklistToken(token string) error {
	// Expiration set to 0, so blacklisting never expires.
	return s.redisClient.BlacklistToken(token, 0)
}

// IsTokenBlacklisted checks whether the given token is blacklisted.
func (s *BlacklistService) IsTokenBlacklisted(token string) (bool, error) {
	blacklisted, err := s.redisClient.IsTokenBlacklisted(token)
	if err != nil {
		return false, err
	}

	return blacklisted, nil
}
