package andromeda

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	endpointGetSites     = "/Sites"
	endpointGetCustomers = "/Customers"
	endpointCheckPanic   = "/CheckPanic"
	endpointMyAlarm      = "/MyAlarm"

	defaultTimeout = 5 * time.Second
)

type (
	//Входная структура для метода GetSites
	GetSitesInput struct {
		Id       int    //Номер объекта
		UserName string //Имя пользователя, от которого делается запрос (необязательное поле)
		ApiKey   string
		Host     string
	}

	//Входная структура для метода GetCustomers
	GetCustomersInput struct {
		SiteId   string //Идентификатор объекта, список ответственных лиц которого нужно получить. Соответствует полю Id карточки объекта
		UserName string //Имя пользователя, от которого делается запрос (необязательное поле)
		ApiKey   string
		Host     string
	}

	//Входная структура для метода GetCustomer
	GetCustomerInput struct {
		Id       string //Идентификатор ответственного лица, информацию которого нужно получить
		UserName string //Имя пользователя, от которого делается запрос (необязательное поле)
		ApiKey   string
		Host     string
	}

	//Входная структура для метода PostCheckPanic
	PostCheckPanicInput struct {
		SiteId        string //Идентификатор объекта, по которому нужно проверить КТС
		CheckInterval int    //Интервал в секундах, в течении которого будет продолжаться процедура проверки КТС. (необязательное поле)
		StopOnEvent   bool   //Признак остановки проверки КТС. (необязательное поле)
		UserName      string //Имя пользователя, от которого делается запрос (необязательное поле)
		ApiKey        string
		Host          string
	}

	//Входная структура для метода GetCheckPanic
	GetCheckPanicInput struct {
		CheckPanicId string //Идентификатор процеудры проверки, для которой нужно получить результат
		UserName     string //Имя пользователя, от которого делается запрос (необязательное поле)
		ApiKey       string
		Host         string
	}

	//Входная структура для метода GetUsersMyAlarm
	GetUsersMyAlarmInput struct {
		SiteId   string //Идентификатор объекта, список пользователей MyAlarm которого нужно получить. Соответствует полю Id карточки объекта
		UserName string //Имя пользователя, от которого делается запрос (необязательное поле)
		ApiKey   string
		Host     string
	}

	request struct {
		URL    string
		body   []byte
		apiKey string
	}

	//Структура для парсинга при ошибке 400 от сервера
	respErr400 struct {
		Message      string `json:"Message"`
		SpResultCode int    `json:"SpResultCode"`
	}

	//Структура ответа от сервера метода PostCheckPanic
	PostCheckPanicResponse struct {
		Status       int    `json:"Status"`
		Description  string `json:"Description"`
		CheckPanicId string `json:"CheckPanicId"`
	}

	//Структура ответа от сервера метода GetCheckPanic
	GetCheckPanicResponse struct {
		Status      int    `json:"Status"`
		Description string `json:"Description"`
	}

	//Структура ответа от сервера метода GetUsersMyAlarm
	UserMyAlarmResponse struct {
		CustomerID   string `json:"CustomerID"`   //Идентификатор пользователя
		MobilePhone  string `json:"MobilePhone"`  //Телефон ответственного
		MyAlarmPhone string `json:"MyAlarmPhone"` //Телефон пользователя MyAlarm
		Role         string `json:"Role"`         //Роль пользователя
		IsPanic      bool   `json:"IsPanic"`      //Разрешён или запрещён КТС
	}

	//Структура ответа от сервера метода GetSites
	GetSitesResponse struct {
		RowNumber                  int     `json:"RowNumber"`                  //Порядковый номер (присутствует только при выводе списка объектов)
		Id                         string  `json:"Id"`                         //Идентификатор объекта
		AccountNumber              int     `json:"AccountNumber"`              //Номер объекта (почти всегда совпадает с номером, запрограммированным в контрольную панель, установленную на объекте)
		CloudObjectID              int     `json:"CloudObjectID"`              //Идентификатор объекта в облаке
		Name                       string  `json:"Name"`                       //Название объекта
		ObjectPassword             string  `json:"ObjectPassword"`             //Пароль объекта
		Address                    string  `json:"Address"`                    //Адрес объекта
		Phone1                     string  `json:"Phone1"`                     //Телефон 1
		Phone2                     string  `json:"Phone2"`                     //Телефон 2
		TypeName                   string  `json:"TypeName"`                   //Название типа объекта
		IsFire                     bool    `json:"IsFire"`                     //Флаг наличия пожарной сигнализации на объекте
		IsArm                      bool    `json:"IsArm"`                      //Флаг наличия охранной сигнализации на объекте
		IsPanic                    bool    `json:"IsPanic"`                    //Флаг наличия тревожной кнопки на объекте
		DeviceTypeName             string  `json:"DeviceTypeName"`             //Псевдоним типа оборудования на объекте.
		EventTemplateName          string  `json:"EventTemplateName"`          //Название шаблона событий объекта
		ContractNumber             string  `json:"ContractNumber"`             //Номер договора
		ContractPrice              float64 `json:"ContractPrice"`              //Сумма ежемесячного платежа по договору. Отображается в приложении MyAlarm
		MoneyBalance               float64 `json:"MoneyBalance"`               //Баланс лицевого счета. Отображается в приложении MyAlarm
		PaymentDate                string  `json:"PaymentDate"`                //Дата ближайшего списания средств. Отображается в приложении	MyAlarm
		DebtInformLevel            int     `json:"DebtInformLevel"`            //Уровень информирования клиента о состоянии услуг охраны. Отображается в приложении MyAlarm.
		Disabled                   bool    `json:"Disabled"`                   //Флаг: объект отключен
		DisableReason              int     `json:"DisableReason"`              //Код: причина отключения объекта (не используется)
		DisableDate                string  `json:"DisableDate"`                //Дата отключения объекта
		AutoEnable                 bool    `json:"AutoEnable"`                 //Флаг: необходимо автоматически включить объект
		AutoEnableDate             string  `json:"AutoEnableDate"`             //Дата автоматического включения объекта. Имеет значение только в том случае, если поле «AutoEnable» установлено в значение «True»
		CustomersComment           string  `json:"CustomersComment"`           //Комментарий к списку ответственных
		CommentForOperator         string  `json:"CommentForOperator"`         //Комментарий для оператора
		CommentForGuard            string  `json:"CommentForGuard"`            //Комментарий для ГБР
		MapFileName                string  `json:"MapFileName"`                //Путь к файлу с картой объекта
		WebLink                    string  `json:"WebLink"`                    //Web-ссылка: ссылка на ресурс с дополнительной информацией об объекте
		ControlTime                int     `json:"ControlTime"`                //Общее контрольное время (мин.)
		CTIgnoreSystemEvent        bool    `json:"CTIgnoreSystemEvent"`        //Игнорировать системные события
		IsContractPriceForceUpdate bool    `json:"IsContractPriceForceUpdate"` //Признак принудительной записи поля ContractPrice
		IsMoneyBalanceForceUpdate  bool    `json:"IsMoneyBalanceForceUpdate"`  //Признак принудительной записи поля MoneyBalance
		IsPaymentDateForceUpdate   bool    `json:"IsPaymentDateForceUpdate"`   //Признак принудительной записи поля PaymentDate
		IsStateArm                 bool    `json:"IsStateArm"`                 //Состояние объекта: взят/снят/неизвестно.
		IsStateAlarm               bool    `json:"IsStateAlarm"`               //Состояние объекта: объект в тревоге - да/нет.
		IsStatePartArm             bool    `json:"IsStatePartArm"`             //Состояние объекта: частично - да/нет/неизвестно.
		StateArmDisArmDateTime     string  `json:"StateArmDisArmDateTime"`     //Состояние объекта: время последнего взятия / снятия.
	}

	//Структура ответа метода GetCustomers, GetCustomer
	GetCustomerResponse struct {
		Id                 string `json:"Id"`                 //Идентификатор ответственного лица
		OrderNumber        int    `json:"OrderNumber"`        //Порядковый номер ответственного в списк (уникальный на объекте, может быть не задан)
		UserNumber         int    `json:"UserNumber"`         //Номер ответственного (номер пользователя на контрольной панели, натуральное число, уникальный на объекте, может быть не задан, нельзя очистить для пользователя MyAlarm)
		ObjCustName        string `json:"ObjCustName"`        //ФИО
		ObjCustTitle       string `json:"ObjCustTitle"`       //Должность
		ObjCustPhone1      string `json:"ObjCustPhone1"`      //Мобильный телефон (уникальный на объекте, нельзя изменить для пользователя MyAlarm)
		ObjCustPhone2      string `json:"ObjCustPhone2"`      //Телефон 2
		ObjCustPhone3      string `json:"ObjCustPhone3"`      //Телефон 3
		ObjCustPhone4      string `json:"ObjCustPhone4"`      //Телефон 4
		ObjCustPhone5      string `json:"ObjCustPhone5"`      //Телефон 5
		ObjCustAddress     string `json:"ObjCustAddress"`     //Адрес
		IsVisibleInCabinet bool   `json:"IsVisibleInCabinet"` //Отображать в личном кабинете (нельзя отключить для пользователя	MyAlarm)
		ReclosingRequest   bool   `json:"ReclosingRequest"`   //Отправлять SMS о необходимости перезакрытия
		ReclosingFailure   bool   `json:"ReclosingFailure"`   //Отправлять SMS об отказе от перезакрытия
		PINCode            string `json:"PINCode"`            //PIN для Call-центра
	}
)

// Проверка заполнения обязательных полей метода GetSites
func (i GetSitesInput) validate() error {
	if i.Id < 1 {
		return errors.New("неверно задан номер объекта")
	}

	if i.ApiKey == "" {
		return errors.New("неверно задан API ключ")
	}

	if i.Host == "" {
		return errors.New("неверно задан адрес сервера")
	}

	return nil
}

// Проверка заполнения обязательных полей метода GetCustomers
func (i GetCustomersInput) validate() error {
	if i.SiteId == "" {
		return errors.New("неверно задан номер объекта")
	}

	if i.ApiKey == "" {
		return errors.New("неверно задан API ключ")
	}

	if i.Host == "" {
		return errors.New("неверно задан адрес сервера")
	}

	return nil
}

// Проверка заполнения обязательных полей метода GetCustomer
func (i GetCustomerInput) validate() error {
	if i.Id == "" {
		return errors.New("неверно задан идентификатор ответственного лица")
	}

	if i.ApiKey == "" {
		return errors.New("неверно задан API ключ")
	}

	if i.Host == "" {
		return errors.New("неверно задан адрес сервера")
	}

	return nil
}

// Проверка заполнения обязательных полей метода PostCheckPanic
func (i PostCheckPanicInput) validate() error {
	if i.SiteId == "" {
		return errors.New("неверно задан номер объекта")
	}

	if i.ApiKey == "" {
		return errors.New("неверно задан API ключ")
	}

	if i.Host == "" {
		return errors.New("неверно задан адрес сервера")
	}

	if i.CheckInterval != 0 {
		if i.CheckInterval <= 30 || i.CheckInterval >= 180 {
			return errors.New("неверно задано время ожидания проверки")
		}
	}
	return nil
}

// Проверка заполнения обязательных полей метода PostCheckPanic
func (i GetCheckPanicInput) validate() error {
	if i.CheckPanicId == "" {
		return errors.New("неверно задан идентификатор проверки")
	}

	if i.ApiKey == "" {
		return errors.New("неверно задан API ключ")
	}

	if i.Host == "" {
		return errors.New("неверно задан адрес сервера")
	}

	return nil
}

// Проверка заполнения обязательных полей метода PostCheckPanic
func (i GetUsersMyAlarmInput) validate() error {
	if i.SiteId == "" {
		return errors.New("неверно задан идентификатор проверки")
	}

	if i.ApiKey == "" {
		return errors.New("неверно задан API ключ")
	}

	if i.Host == "" {
		return errors.New("неверно задан адрес сервера")
	}

	return nil
}

// Генерация запроса метода GetSites
func (i GetSitesInput) generateRequest() request {
	baseURL, _ := url.Parse(i.Host + endpointGetSites)
	param := url.Values{}
	param.Add("id", strconv.Itoa(i.Id))
	if i.UserName != "" {
		param.Add("userName", i.UserName)
	}
	baseURL.RawQuery = param.Encode()

	return request{
		URL:    baseURL.String(),
		body:   []byte{},
		apiKey: i.ApiKey,
	}
}

// Генерация запроса метода GetCustomers
func (i GetCustomersInput) generateRequest() request {

	baseURL, _ := url.Parse(i.Host + endpointGetCustomers)
	param := url.Values{}
	param.Add("siteId", i.SiteId)
	if i.UserName != "" {
		param.Add("userName", i.UserName)
	}
	baseURL.RawQuery = param.Encode()

	return request{
		URL:    baseURL.String(),
		body:   []byte{},
		apiKey: i.ApiKey,
	}

}

// Генерация запроса метода GetCustomer
func (i GetCustomerInput) generateRequest() request {

	baseURL, _ := url.Parse(i.Host + endpointGetCustomers)
	param := url.Values{}
	param.Add("id", i.Id)
	if i.UserName != "" {
		param.Add("userName", i.UserName)
	}
	baseURL.RawQuery = param.Encode()

	return request{
		URL:    baseURL.String(),
		body:   []byte{},
		apiKey: i.ApiKey,
	}

}

// Генерация запроса метода PostCheckPanic
func (i PostCheckPanicInput) generateRequest() request {

	baseURL, _ := url.Parse(i.Host + endpointCheckPanic)
	param := url.Values{}
	param.Add("siteId", i.SiteId)
	param.Add("stopOnEvent", "True")
	if i.CheckInterval != 0 {
		param.Add("checkInterval", strconv.Itoa(i.CheckInterval))
	}
	if i.UserName != "" {
		param.Add("userName", i.UserName)
	}
	baseURL.RawQuery = param.Encode()

	return request{
		URL:    baseURL.String(),
		body:   []byte{},
		apiKey: i.ApiKey,
	}

}

// Генерация запроса метода GetCheckPanic
func (i GetCheckPanicInput) generateRequest() request {

	baseURL, _ := url.Parse(i.Host + endpointCheckPanic)
	param := url.Values{}
	param.Add("checkPanicId", i.CheckPanicId)
	if i.UserName != "" {
		param.Add("userName", i.UserName)
	}
	baseURL.RawQuery = param.Encode()

	return request{
		URL:    baseURL.String(),
		body:   []byte{},
		apiKey: i.ApiKey,
	}

}

func (i GetUsersMyAlarmInput) generateRequest() request {

	baseURL, _ := url.Parse(i.Host + endpointMyAlarm)
	param := url.Values{}
	param.Add("siteId", i.SiteId)
	if i.UserName != "" {
		param.Add("userName", i.UserName)
	}
	baseURL.RawQuery = param.Encode()

	return request{
		URL:    baseURL.String(),
		body:   []byte{},
		apiKey: i.ApiKey,
	}

}

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{Timeout: defaultTimeout},
	}
}

// Запрос метода GetSites
func (c *Client) GetSites(ctx context.Context, input GetSitesInput) (GetSitesResponse, error) {
	if err := input.validate(); err != nil {
		return GetSitesResponse{}, err
	}

	req := input.generateRequest()
	body, err := c.doHTTP(ctx, http.MethodGet, req)
	if err != nil {
		return GetSitesResponse{}, err
	}

	var resp GetSitesResponse

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return GetSitesResponse{}, errors.WithMessage(err, "Не удалось парсить ответ")
	}

	return resp, nil
}

// Запрос метода GetCustomers
func (c *Client) Customers(ctx context.Context, input GetCustomersInput) ([]GetCustomerResponse, error) {
	if err := input.validate(); err != nil {
		return []GetCustomerResponse{}, err
	}

	req := input.generateRequest()
	body, err := c.doHTTP(ctx, http.MethodGet, req)
	if err != nil {
		return []GetCustomerResponse{}, err
	}

	resp := []GetCustomerResponse{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return []GetCustomerResponse{}, errors.WithMessage(err, "Не удалось парсить ответ")
	}

	return resp, nil
}

// Запрос метода GetCustomer
func (c *Client) GetCustomer(ctx context.Context, input GetCustomerInput) (GetCustomerResponse, error) {
	if err := input.validate(); err != nil {
		return GetCustomerResponse{}, err
	}

	req := input.generateRequest()
	body, err := c.doHTTP(ctx, http.MethodGet, req)
	if err != nil {
		return GetCustomerResponse{}, err
	}

	resp := GetCustomerResponse{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return GetCustomerResponse{}, errors.WithMessage(err, "Не удалось парсить ответ")
	}

	return resp, nil
}

// Запрос метода PostCheckPanic
func (c *Client) PostCheckPanic(ctx context.Context, input PostCheckPanicInput) (PostCheckPanicResponse, error) {
	if err := input.validate(); err != nil {
		return PostCheckPanicResponse{}, err
	}

	req := input.generateRequest()
	body, err := c.doHTTP(ctx, http.MethodPost, req)
	if err != nil {
		return PostCheckPanicResponse{}, err
	}

	var resp PostCheckPanicResponse

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return PostCheckPanicResponse{}, errors.WithMessage(err, "Не удалось парсить ответ")
	}

	return resp, nil
}

// Запрос метода GetCheckPanic
func (c *Client) GetCheckPanic(ctx context.Context, input GetCheckPanicInput) (GetCheckPanicResponse, error) {
	if err := input.validate(); err != nil {
		return GetCheckPanicResponse{}, err
	}

	req := input.generateRequest()
	body, err := c.doHTTP(ctx, http.MethodGet, req)
	if err != nil {
		return GetCheckPanicResponse{}, err
	}

	var resp GetCheckPanicResponse

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return GetCheckPanicResponse{}, errors.WithMessage(err, "Не удалось парсить ответ")
	}

	return resp, nil
}

// Запрос метода GetUsersMyAlarm
func (c *Client) GetUsersMyAlarm(ctx context.Context, input GetUsersMyAlarmInput) ([]UserMyAlarmResponse, error) {
	if err := input.validate(); err != nil {
		return []UserMyAlarmResponse{}, err
	}

	req := input.generateRequest()
	body, err := c.doHTTP(ctx, http.MethodGet, req)
	if err != nil {
		return []UserMyAlarmResponse{}, err
	}

	var resp []UserMyAlarmResponse

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return []UserMyAlarmResponse{}, errors.WithMessage(err, "Не удалось парсить ответ")
	}

	return resp, nil
}

// Метод http выполнения запроса
func (c *Client) doHTTP(ctx context.Context, method string, r request) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, method, r.URL, bytes.NewBuffer(r.body))
	if err != nil {
		return []byte{}, errors.WithMessage(err, "Не удалось создать запрос")
	}

	req.Header.Set("apiKey", r.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return []byte{}, errors.WithMessage(err, "Не удалось выполнить запрос")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, resp.Body); err != nil {
			return []byte{}, errors.WithMessage(err, "Не удалось выполнить запрос")
		}
		err400 := respErr400{}
		err = json.Unmarshal(buf.Bytes(), &err400)
		if err != nil {
			return []byte{}, err
		}
		return []byte{}, errors.New(err400.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return []byte{}, errors.New("Не удалось выполнить запрос")
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return []byte{}, errors.WithMessage(err, "Не удалcя парсинг ответа")
	}

	return buf.Bytes(), nil
}
