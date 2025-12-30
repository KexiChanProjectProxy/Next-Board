package repository

import (
	"time"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"

	"gorm.io/gorm"
)

type NodeRepository interface {
	Create(node *models.Node) error
	FindByID(id uint64) (*models.Node, error)
	Update(node *models.Node) error
	Delete(id uint64) error
	List(offset, limit int) ([]models.Node, int64, error)
	FindByIDWithLabels(id uint64) (*models.Node, error)
	FindActiveNodes() ([]models.Node, error)
	UpdateLastSeen(nodeID uint64) error
	AddLabel(nodeID, labelID uint64) error
	RemoveLabel(nodeID, labelID uint64) error
	GetLabels(nodeID uint64) ([]models.Label, error)
}

type nodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepository{db: db}
}

func (r *nodeRepository) Create(node *models.Node) error {
	return r.db.Create(node).Error
}

func (r *nodeRepository) FindByID(id uint64) (*models.Node, error) {
	var node models.Node
	err := r.db.First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *nodeRepository) FindByIDWithLabels(id uint64) (*models.Node, error) {
	var node models.Node
	err := r.db.Preload("Labels").First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *nodeRepository) Update(node *models.Node) error {
	return r.db.Save(node).Error
}

func (r *nodeRepository) Delete(id uint64) error {
	return r.db.Delete(&models.Node{}, id).Error
}

func (r *nodeRepository) List(offset, limit int) ([]models.Node, int64, error) {
	var nodes []models.Node
	var total int64

	if err := r.db.Model(&models.Node{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Labels").Offset(offset).Limit(limit).Find(&nodes).Error
	return nodes, total, err
}

func (r *nodeRepository) FindActiveNodes() ([]models.Node, error) {
	var nodes []models.Node
	err := r.db.Preload("Labels").Where("status = ?", "active").Find(&nodes).Error
	return nodes, err
}

func (r *nodeRepository) UpdateLastSeen(nodeID uint64) error {
	now := time.Now()
	return r.db.Model(&models.Node{}).Where("id = ?", nodeID).Update("last_seen_at", now).Error
}

func (r *nodeRepository) AddLabel(nodeID, labelID uint64) error {
	nodeLabel := &models.NodeLabel{
		NodeID:  nodeID,
		LabelID: labelID,
	}
	return r.db.Create(nodeLabel).Error
}

func (r *nodeRepository) RemoveLabel(nodeID, labelID uint64) error {
	return r.db.Where("node_id = ? AND label_id = ?", nodeID, labelID).Delete(&models.NodeLabel{}).Error
}

func (r *nodeRepository) GetLabels(nodeID uint64) ([]models.Label, error) {
	var node models.Node
	err := r.db.Preload("Labels").First(&node, nodeID).Error
	if err != nil {
		return nil, err
	}
	return node.Labels, nil
}
