package usecase

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type AssetIDUsecase struct {
	assetRepo model.IAssetIDRepository
}

func NewAssetIDUsecase(assetRepo model.IAssetIDRepository) model.IAssetIDUsecase {
	return &AssetIDUsecase{assetRepo: assetRepo}
}

func (u *AssetIDUsecase) Create(ctx context.Context, in model.AssetIDInput) (*model.AssetID, error) {
	log := logrus.WithFields(logrus.Fields{"in": in})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return nil, err
	}

	asset := model.AssetID{
		Name:   in.Name,
		PartID: in.PartID,
	}

	created, err := u.assetRepo.Create(ctx, asset)
	if err != nil {
		log.Error("Failed to create asset_id: ", err)
		return nil, err
	}

	return created, nil
}

func (u *AssetIDUsecase) FindAll(ctx context.Context, filter model.AssetID) ([]*model.AssetID, error) {
	log := logrus.WithFields(logrus.Fields{"filter": filter})

	assets, err := u.assetRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("Failed to fetch asset_ids: ", err)
		return nil, err
	}

	return assets, nil
}

func (u *AssetIDUsecase) FindByID(ctx context.Context, id int64) (*model.AssetID, error) {
	log := logrus.WithFields(logrus.Fields{"id": id})

	asset, err := u.assetRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find asset_id: ", err)
		return nil, err
	}

	return asset, nil
}

func (u *AssetIDUsecase) Update(ctx context.Context, id int64, in model.UpdateAssetIDInput) error {
	log := logrus.WithFields(logrus.Fields{"id": id})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return err
	}

	asset, err := u.assetRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if asset == nil {
		return errors.New("asset_id not found")
	}

	asset.Name = in.Name
	asset.PartID = in.PartID

	if err := u.assetRepo.Update(ctx, *asset); err != nil {
		log.Error("Failed to update asset_id: ", err)
		return err
	}

	return nil
}

func (u *AssetIDUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{"id": id})

	asset, err := u.assetRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find asset_id for deletion: ", err)
		return err
	}

	if asset == nil {
		log.Error("Failed to find asset_id for deletion: ", err)
		return errors.New("asset_id not found")
	}

	if err := u.assetRepo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete asset_id: ", err)
		return err
	}

	return nil
}
