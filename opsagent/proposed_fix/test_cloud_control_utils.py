"""
Unit tests for the fixed cloud_control_utils.py module.
These tests verify that the resource-type-aware tagging works correctly.
"""

import pytest
from unittest.mock import patch
from cloud_control_utils import (
    add_default_tags, 
    get_resource_tag_property,
    add_resource_to_tag_map,
    RESOURCE_TAG_PROPERTY_MAP
)


class TestAddDefaultTags:
    """Test cases for the add_default_tags function."""
    
    @patch('cloud_control_utils.get_package_version')
    def test_efs_filesystem_uses_filesystem_tags(self, mock_version):
        """Test that EFS FileSystem resources use FileSystemTags property."""
        mock_version.return_value = "1.0.0"
        
        properties = {
            "PerformanceMode": "generalPurpose",
            "Encrypted": True
        }
        
        result = add_default_tags(properties, {}, "AWS::EFS::FileSystem")
        
        # Should have FileSystemTags, not Tags
        assert "FileSystemTags" in result
        assert "Tags" not in result
        
        # Should have 3 default tags
        filesystem_tags = result["FileSystemTags"]
        assert len(filesystem_tags) == 3
        
        # Verify the default tags are present
        tag_keys = [tag["Key"] for tag in filesystem_tags]
        assert "MANAGED_BY" in tag_keys
        assert "MCP_SERVER_SOURCE_CODE" in tag_keys
        assert "MCP_SERVER_VERSION" in tag_keys
        
        # Verify original properties are preserved
        assert result["PerformanceMode"] == "generalPurpose"
        assert result["Encrypted"] is True
    
    @patch('cloud_control_utils.get_package_version')
    def test_s3_bucket_uses_standard_tags(self, mock_version):
        """Test that S3 Bucket resources use standard Tags property."""
        mock_version.return_value = "1.0.0"
        
        properties = {
            "BucketName": "test-bucket"
        }
        
        result = add_default_tags(properties, {}, "AWS::S3::Bucket")
        
        # Should have Tags, not FileSystemTags
        assert "Tags" in result
        assert "FileSystemTags" not in result
        
        # Should have 3 default tags
        tags = result["Tags"]
        assert len(tags) == 3
        
        # Verify original properties are preserved
        assert result["BucketName"] == "test-bucket"
    
    @patch('cloud_control_utils.get_package_version')
    def test_unknown_resource_type_defaults_to_tags(self, mock_version):
        """Test that unknown resource types default to using Tags property."""
        mock_version.return_value = "1.0.0"
        
        properties = {
            "SomeProperty": "some-value"
        }
        
        result = add_default_tags(properties, {}, "AWS::UnknownService::Resource")
        
        # Should default to Tags
        assert "Tags" in result
        assert len(result["Tags"]) == 3
    
    @patch('cloud_control_utils.get_package_version')
    def test_none_resource_type_defaults_to_tags(self, mock_version):
        """Test that None resource type defaults to using Tags property."""
        mock_version.return_value = "1.0.0"
        
        properties = {
            "SomeProperty": "some-value"
        }
        
        result = add_default_tags(properties, {}, None)
        
        # Should default to Tags
        assert "Tags" in result
        assert len(result["Tags"]) == 3
    
    @patch('cloud_control_utils.get_package_version')
    def test_preserves_existing_tags(self, mock_version):
        """Test that existing tags are preserved and no duplicates are added."""
        mock_version.return_value = "1.0.0"
        
        properties = {
            "FileSystemTags": [
                {"Key": "Environment", "Value": "test"},
                {"Key": "MANAGED_BY", "Value": "existing-value"}  # Should not be overwritten
            ]
        }
        
        result = add_default_tags(properties, {}, "AWS::EFS::FileSystem")
        
        filesystem_tags = result["FileSystemTags"]
        
        # Should have original tags plus 2 new default tags (MANAGED_BY already exists)
        assert len(filesystem_tags) == 4
        
        # Check that existing tags are preserved
        existing_tag = next(tag for tag in filesystem_tags if tag["Key"] == "Environment")
        assert existing_tag["Value"] == "test"
        
        # Check that existing MANAGED_BY is not overwritten
        managed_by_tag = next(tag for tag in filesystem_tags if tag["Key"] == "MANAGED_BY")
        assert managed_by_tag["Value"] == "existing-value"
        
        # Check that other default tags are added
        tag_keys = [tag["Key"] for tag in filesystem_tags]
        assert "MCP_SERVER_SOURCE_CODE" in tag_keys
        assert "MCP_SERVER_VERSION" in tag_keys
    
    def test_empty_properties_returns_empty_dict(self):
        """Test that empty or None properties return empty dict."""
        assert add_default_tags(None, {}, "AWS::S3::Bucket") == {}
        assert add_default_tags({}, {}, "AWS::S3::Bucket") == {}
    
    @patch('cloud_control_utils.get_package_version')
    def test_creates_tag_property_if_missing(self, mock_version):
        """Test that the tag property is created if it doesn't exist."""
        mock_version.return_value = "1.0.0"
        
        properties = {
            "PerformanceMode": "generalPurpose"
            # No FileSystemTags property
        }
        
        result = add_default_tags(properties, {}, "AWS::EFS::FileSystem")
        
        # Should create FileSystemTags property
        assert "FileSystemTags" in result
        assert len(result["FileSystemTags"]) == 3


class TestGetResourceTagProperty:
    """Test cases for the get_resource_tag_property function."""
    
    def test_efs_filesystem_returns_filesystem_tags(self):
        """Test that EFS FileSystem returns FileSystemTags."""
        result = get_resource_tag_property("AWS::EFS::FileSystem")
        assert result == "FileSystemTags"
    
    def test_s3_bucket_returns_tags(self):
        """Test that S3 Bucket returns Tags."""
        result = get_resource_tag_property("AWS::S3::Bucket")
        assert result == "Tags"
    
    def test_unknown_resource_returns_tags(self):
        """Test that unknown resource types return Tags."""
        result = get_resource_tag_property("AWS::Unknown::Resource")
        assert result == "Tags"


class TestAddResourceToTagMap:
    """Test cases for the add_resource_to_tag_map function."""
    
    def test_adds_new_resource_type(self):
        """Test that new resource types can be added to the mapping."""
        # Save original state
        original_map = RESOURCE_TAG_PROPERTY_MAP.copy()
        
        try:
            # Add new resource type
            add_resource_to_tag_map("AWS::NewService::Resource", "ResourceTags")
            
            # Verify it was added
            assert get_resource_tag_property("AWS::NewService::Resource") == "ResourceTags"
            
        finally:
            # Restore original state
            RESOURCE_TAG_PROPERTY_MAP.clear()
            RESOURCE_TAG_PROPERTY_MAP.update(original_map)
    
    def test_overwrites_existing_resource_type(self):
        """Test that existing resource types can be overwritten."""
        # Save original state
        original_map = RESOURCE_TAG_PROPERTY_MAP.copy()
        
        try:
            # Overwrite existing resource type
            add_resource_to_tag_map("AWS::EFS::FileSystem", "NewTagProperty")
            
            # Verify it was overwritten
            assert get_resource_tag_property("AWS::EFS::FileSystem") == "NewTagProperty"
            
        finally:
            # Restore original state
            RESOURCE_TAG_PROPERTY_MAP.clear()
            RESOURCE_TAG_PROPERTY_MAP.update(original_map)


if __name__ == "__main__":
    pytest.main([__file__])