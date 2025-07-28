package databasedriver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Blanco0420/Phone-Number-Check/backend/config"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/address"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/business"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/caller"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/carrier"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/linetype"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/number"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/provider"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	_ "github.com/lib/pq"
)

type DatabaseDriver struct {
	client *ent.Client
}

func readSecret(path string) (string, error) {
	var content string
	rawContent, err := os.ReadFile(path)
	if err != nil {
		return content, fmt.Errorf("error reading secrets file %s: %v", path, err)
	}
	content = strings.TrimSpace(string(rawContent))
	return content, nil
}

func InitializeDriver() (*DatabaseDriver, error) {
	const (
		defaultHost = "phraud-database"
		defaultPort = "5432"
		defaultUser = "postgres"
	)
	getEnvOrDefault := func(key, defaultValue string) string {
		val, exists := config.GetEnvVariable(key)
		if !exists {
			return defaultValue
		}
		return val
	}

	host := getEnvOrDefault("PHRAUD__DB_HOST", defaultHost)
	port := getEnvOrDefault("PHRAUD__DB_PORT", defaultPort)
	user := getEnvOrDefault("PHRAUD__DB_USER", defaultUser)
	var password string
	password, exists := config.GetEnvVariable("PHRAUD__DB_PASS")
	if !exists {
		passwordFile, exists := config.GetEnvVariable("PHRAUD__DB_PASS_FILE")
		if !exists {
			err := errors.New("failed to find password and password file")
			return nil, err
		}
		var err error
		password, err = readSecret(passwordFile)
		if err != nil {
			return nil, err
		}

	}
	client, err := ent.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=Phraud password=%s sslmode=disable", host, port, user, password))
	if err != nil {
		return nil, err
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed creating schema resources: %v", err)
	}
	return &DatabaseDriver{client: client}, nil
}

func getCommentBuilders(tx *ent.Tx, comments []providers.Comment) []*ent.CommentCreate {
	builders := make([]*ent.CommentCreate, 0, len(comments))
	for _, comment := range comments {
		builder := tx.Comment.Create().
			SetCommentText(comment.Text).
			SetPostDate(comment.PostDate)
		builders = append(builders, builder)
	}

	return builders
}

func getOrCreate[T any](queryFn func() (*T, error), createFn func() (*T, error)) (*T, error) {
	res, err := queryFn()
	if err == nil {
		return res, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	return createFn()
}

func (d *DatabaseDriver) InsertNumberIntoDatabase(ctx context.Context, data map[string]providers.NumberDetails, finalFraudScore int) error {
	tx, err := d.client.Tx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
		if !committed {
			_ = tx.Rollback()
		}
	}()

	for providerName, details := range data {

		// Try to find or create Carrier
		// carrier, err := tx.Carrier.Query().
		// 	Where(carrier.NameEQ(*details.Carrier)).
		// 	Only(ctx)
		// if ent.IsNotFound(err) {
		// 	carrier, err = tx.Carrier.Create().
		// 		SetName(*details.Carrier).
		// 		Save(ctx)
		// }
		// if err != nil {
		// 	return err
		// }
		var carrierSchema *ent.Carrier
		if details.Carrier != nil {
			carrierSchema, err = getOrCreate(
				func() (*ent.Carrier, error) {
					return tx.Carrier.Query().Where(carrier.NameEQ(*details.Carrier)).Only(ctx)
				},

				func() (*ent.Carrier, error) {
					return tx.Carrier.Create().
						SetNillableName(details.Carrier).
						Save(ctx)
				},
			)
			if err != nil {
				return err
			}
		}

		lineType, err := getOrCreate(
			func() (*ent.LineType, error) {
				return tx.LineType.Query().Where(linetype.LineTypeEQ(details.VitalInfo.LineType)).Only(ctx)
			},
			func() (*ent.LineType, error) {
				return tx.LineType.Create().
					SetLineType(details.VitalInfo.LineType).
					Save(ctx)
			},
		)
		if err != nil {
			return err
		}

		caller, err := getOrCreate(
			func() (*ent.Caller, error) {
				return tx.Caller.Query().Where(caller.HasNumberWith(number.NumberEQ(details.Number))).Only(ctx)
			},
			func() (*ent.Caller, error) {
				return tx.Caller.Create().
					SetFraudScore(finalFraudScore).
					SetIsFraud(finalFraudScore > 40).
					Save(ctx)
			},
		)
		if err != nil {
			return err
		}

		number, err := getOrCreate(
			func() (*ent.Number, error) {
				return tx.Number.Query().Where(number.NumberEQ(details.Number)).
					Only(ctx)
			},
			func() (*ent.Number, error) {
				return tx.Number.Create().
					SetNumber(details.Number).
					SetCarrier(carrierSchema).
					SetCaller(caller).
					SetLinetype(lineType).
					Save(ctx)
			},
		)
		if err != nil {
			return err
		}

		provider, err := getOrCreate(
			func() (*ent.Provider, error) {
				return tx.Provider.Query().Where(provider.NameEQ(providerName)).Only(ctx)
			},

			func() (*ent.Provider, error) {
				return tx.Provider.Create().
					SetName(providerName).
					SetNumber(number).
					Save(ctx)
			},
		)
		if err != nil {
			return err
		}

		// Try to find or create LineType
		// lineType, err := tx.LineType.Query().
		// 	Where(linetype.LineTypeEQ(details.VitalInfo.LineType)).
		// 	Only(ctx)
		// if ent.IsNotFound(err) {
		// 	lineType, err = tx.LineType.Create().
		// 		SetLineType(details.VitalInfo.LineType).
		// 		Save(ctx)
		// }
		// if err != nil {
		// 	return err
		// }

		// Always create new Caller (assumed unique for each request)
		// caller, err := tx.Caller.Create().
		// 	SetFraudScore(finalFraudScore).
		// 	SetIsFraud(finalFraudScore > 40).
		// 	Save(ctx)
		// if err != nil {
		// 	return err
		// }

		// Always create new Business
		var businessSchema *ent.Business
		if details.BusinessDetails != nil {
			businessSchema, err = getOrCreate(
				func() (*ent.Business, error) {
					return tx.Business.Query().Where(business.NameEQ(safeString(details.VitalInfo.Name))).Only(ctx)
				},
				func() (*ent.Business, error) {
					return tx.Business.Create().
						SetNillableName(details.VitalInfo.Name).
						SetNillableOverview(details.VitalInfo.CompanyOverview).
						SetProvider(provider).
						SetNillableWebsiteLink(details.BusinessDetails.Website).
						Save(ctx)
				},
			)
			if err != nil {
				return err
			}
			if details.BusinessDetails.LocationDetails != nil {
				if details.BusinessDetails.LocationDetails.PostCode != nil {
					_, err = getOrCreate(
						func() (*ent.Address, error) {
							return tx.Address.Query().
								Where(address.PostcodeEQ(safeString(details.BusinessDetails.LocationDetails.PostCode))).
								Only(ctx)
						},
						func() (*ent.Address, error) {
							return tx.Address.Create().
								SetBusiness(businessSchema).
								SetNillableCity(details.BusinessDetails.LocationDetails.City).
								SetNillablePrefecture(details.BusinessDetails.LocationDetails.Prefecture).
								SetNillablePostcode(details.BusinessDetails.LocationDetails.PostCode).
								Save(ctx)
						},
					)
					if err != nil {
						return err
					}
				} else if details.BusinessDetails.LocationDetails.City != nil {
					_, err := tx.Address.Query().
						Where(address.CityEQ(safeString(details.BusinessDetails.LocationDetails.City))).
						Only(ctx)
					if !ent.IsNotFound(err) {
						return err
					}
				}
			}
		}
		// tx.Business.Create().
		// 	SetName(*details.VitalInfo.Name).
		// 	SetOverview(*details.VitalInfo.CompanyOverview).
		// 	SetProvider(provider).
		// 	Save(ctx)
		// if err != nil {
		// 	return err
		// }

		// Always create new Address

		// Bulk insert Comments
		commentBuilders := getCommentBuilders(tx, details.SiteInfo.Comments)
		if _, err := tx.Comment.CreateBulk(commentBuilders...).Save(ctx); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (d *DatabaseDriver) Close() {
	if d.client != nil {
		_ = d.client.Close()
	}
}

// safeString returns the dereferenced value of a string pointer, or "" if nil.
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
