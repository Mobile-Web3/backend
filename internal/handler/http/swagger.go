package http

// CreateMnemonic godoc
// @Summary      Создание мнемоника
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.CreateMnemonicInput true "body"
// @Success      200 {object} apiResponse{result=string}
// @Router       /account/mnemonic [post]
func CreateMnemonic() {}

// CreateAccount godoc
// @Summary      Получение аккаунта по мнемонику
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.CreateAccountInput true "body"
// @Success      200 {object} apiResponse{result=chain.AccountResponse}
// @Router       /account/create [post]
func CreateAccount() {}

// RestoreAccount godoc
// @Summary      Получение аккаунта по ключу
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.RestoreAccountInput true "body"
// @Success      200 {object} apiResponse{result=chain.AccountResponse}
// @Router       /account/restore [post]
func RestoreAccount() {}

// CheckBalance godoc
// @Summary      Получить инфу о балансе
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.BalanceInput true "body"
// @Success      200 {object} apiResponse{result=chain.BalanceResponse}
// @Router       /account/balance [post]
func CheckBalance() {}

// GetAllChains godoc
// @Summary      Получение данных о сетях
// @Tags         chain
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @Success      200 {object} apiResponse{result=[]chain.ShortResponse}
// @Router       /chains/all [post]
func GetAllChains() {}

// SendTransaction godoc
// @Summary      Отправить транзакцию
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.SendTxInput true "body"
// @Success      200 {object} apiResponse{result=chain.SendTxResponse}
// @Router       /transaction/send [post]
func SendTransaction() {}

// SimulateTransaction godoc
// @Summary      Симуляция транзакции для расчета параметров
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.SimulateTxInput true "body"
// @Success      200 {object} apiResponse{result=chain.SimulateTxResponse}
// @Router       /transaction/simulate [post]
func SimulateTransaction() {}
