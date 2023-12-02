//rodou mas precisacolcoar pra retornar os valores da querybanco no get all assets. Será que da pra chamar a query do banco?
package chaincode

import (
        "database/sql"
        "encoding/json"
        "fmt"

        _ "github.com/go-sql-driver/mysql"
        "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// MyContract é o contrato inteligente para o Hyperledger Fabric
type MyContract struct {
        contractapi.Contract
        Assets []*TripData // Declare a variável 'assets' como um campo da estrutura
}

// TripData estrutura para representar os dados de uma viagem
type TripData struct {
        ID                string  `json:"ID"`
        DepartureDatetime string  `json:"Departure_Datetime"`
        TotalDistanceKm   float64 `json:"totalDistance_km"`
        TripID            int     `json:"TripID"`
        ArrivalDatetime   string  `json:"Arrival_Datetime"`
}

// GetAllAssets retorna todos os ativos encontrados no estado mundial
func (mc *MyContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*TripData, error) {
        resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
        if err != nil {
                return nil, fmt.Errorf("falha ao obter ativos: %v", err)
        }
        defer resultsIterator.Close()

        var assets []*TripData
        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return nil, fmt.Errorf("falha ao iterar sobre resultados de consulta: %v", err)
                }

                var asset TripData
                err = json.Unmarshal(queryResponse.Value, &asset)
                if err != nil {
                        return nil, fmt.Errorf("falha ao fazer unmarshal dos dados de viagem: %v", err)
                }
                assets = append(assets, &asset)
        }

        return assets, nil
}

// InitLedger inicializa o estado mundial com dados de uma consulta SQL
func (mc *MyContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
        mc.Assets = []*TripData{
                {ID: "asset1", DepartureDatetime: "blue", TotalDistanceKm: 5, TripID: 1, ArrivalDatetime: "test"},
                {ID: "asset2", DepartureDatetime: "red", TotalDistanceKm: 8, TripID: 2, ArrivalDatetime: "sample"},
        }
        for _, asset := range mc.Assets {
                assetJSON, err := json.Marshal(asset)
                if err != nil {
                        return err
                }

                err = ctx.GetStub().PutState(asset.ID, assetJSON)
                if err != nil {
                        return fmt.Errorf("failed to put to world state. %v", err)
                }
        }

        return nil
}

func (mc *MyContract) QueryDatabase(ctx contractapi.TransactionContextInterface) error {
        db, err := sql.Open("mysql", "root:movepass@tcp(192.168.10.24:3306)/moveuff")
        if err != nil {
                return fmt.Errorf("falha ao conectar ao banco de dados: %v", err)
        }
        defer db.Close()

        queryToday := `
        SELECT
            departure.id AS Departure_Datetime,
            trips.totalDistance_km,
            trips.id AS TripID,
            arrival.id AS Arrival_Datetime
        FROM trip_x_parkingslot_departures AS departure
        JOIN trips ON departure.Trips_id = trips.id
        JOIN trip_x_parkingslot_arrivals AS arrival ON arrival.Trips_id = trips.id
        WHERE DATE(departure.id) = CURDATE() AND DATE(arrival.id) = CURDATE()
    `

        rows, err := db.Query(queryToday)
        if err != nil {
                return fmt.Errorf("falha ao executar a query: %v", err)
        }
        defer rows.Close()

        var tripDataList []map[string]interface{}

        for rows.Next() {
                var departureDatetime string
                var totalDistanceKm float64
                var tripID int
                var arrivalDatetime string

                err := rows.Scan(&departureDatetime, &totalDistanceKm, &tripID, &arrivalDatetime)
                if err != nil {
                        return fmt.Errorf("falha ao ler os valores do resultado: %v", err)
                }

                rowData := map[string]interface{}{
                        "Departure_Datetime": departureDatetime,
                        "totalDistance_km":   totalDistanceKm,
                        "TripID":             tripID,
                        "Arrival_Datetime":   arrivalDatetime,
                }

                tripDataList = append(tripDataList, rowData)
        }

        for _, data := range tripDataList {
                asset := &TripData{
                        ID:                fmt.Sprintf("asset%d", len(mc.Assets)+1),
                        DepartureDatetime: data["Departure_Datetime"].(string),
                        TotalDistanceKm:   data["totalDistance_km"].(float64),
                        TripID:            data["TripID"].(int),
                        ArrivalDatetime:   data["Arrival_Datetime"].(string),
                }
                mc.Assets = append(mc.Assets, asset)
        }

        for _, asset := range mc.Assets {
                assetJSON, err := json.Marshal(asset)
                if err != nil {
                        return fmt.Errorf("falha ao converter ativo para JSON: %v", err)
                }

                err = ctx.GetStub().PutState(asset.ID, assetJSON)
                if err != nil {
                        return fmt.Errorf("falha ao colocar no estado mundial: %v", err)
                }
        }

        return nil
}

// ReadTripData retorna os dados de viagem armazenados no estado mundial com o ID fornecido.
func (mc *MyContract) ReadTripData(ctx contractapi.TransactionContextInterface, id string) (*TripData, error) {
        tripDataJSON, err := ctx.GetStub().GetState(id)
        if err != nil {
                return nil, fmt.Errorf("falha ao ler do estado mundial: %v", err)
        }
        if tripDataJSON == nil {
                return nil, fmt.Errorf("os dados de viagem %s não existem", id)
        }

        var tripData TripData
        err = json.Unmarshal(tripDataJSON, &tripData)
        if err != nil {
                return nil, fmt.Errorf("falha ao fazer unmarshal dos dados de viagem: %v", err)
        }

        return &tripData, nil
}

// DeleteTripData exclui dados de viagem fornecidos do estado mundial.
func (mc *MyContract) DeleteTripData(ctx contractapi.TransactionContextInterface, id string) error {
        exists, err := mc.TripDataExists(ctx, id)
        if err != nil {
                return fmt.Errorf("falha ao verificar a existência de dados de viagem: %v", err)
        }
        if !exists {
                return fmt.Errorf("os dados de viagem %s não existem", id)
        }

        return ctx.GetStub().DelState(id)
}

// TripDataExists retorna true quando dados de viagem com o ID fornecido existem no estado mundial.
func (mc *MyContract) TripDataExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
        tripDataJSON, err := ctx.GetStub().GetState(id)
        if err != nil {
                return false, fmt.Errorf("falha ao ler do estado mundial: %v", err)
        }

        return tripDataJSON != nil, nil
}

// GetAllTripData retorna todos os dados de viagem encontrados no estado mundial.
func (mc *MyContract) GetAllTripData(ctx contractapi.TransactionContextInterface) ([]*TripData, error) {
        resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
        if err != nil {
                return nil, fmt.Errorf("falha ao obter dados de viagem: %v", err)
        }
        defer resultsIterator.Close()

        var tripDataList []*TripData
        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return nil, fmt.Errorf("falha ao iterar sobre resultados de consulta: %v", err)
                }

                var tripData TripData
                err = json.Unmarshal(queryResponse.Value, &tripData)
                if err != nil {
                        return nil, fmt.Errorf("falha ao fazer unmarshal dos dados de viagem: %v", err)
                }
                tripDataList = append(tripDataList, &tripData)
        }

        return tripDataList, nil
}
