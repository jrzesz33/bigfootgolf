"""
Fixed version of cloud_control_utils.py with resource-type-aware tagging
This addresses the EFS tagging issue where 'Tags' is not permitted.
"""

from typing import Dict, Any, List
from awslabs.ccapi_mcp_server.utils import get_package_version

# Resource types that use non-standard tag properties
RESOURCE_TAG_PROPERTY_MAP = {
    'AWS::EFS::FileSystem': 'FileSystemTags',
    'AWS::EFS::AccessPoint': 'Tags',  # EFS Access Points use Tags
    'AWS::CloudFormation::Stack': 'Tags',
    'AWS::ECR::Repository': 'Tags',
    'AWS::RDS::DBInstance': 'Tags',
    'AWS::S3::Bucket': 'Tags',
    'AWS::EC2::Instance': 'Tags',
    'AWS::ECS::Cluster': 'Tags',
    'AWS::ECS::TaskDefinition': 'Tags',
    'AWS::ECS::Service': 'Tags',
    'AWS::SecretsManager::Secret': 'Tags',
    'AWS::ElasticLoadBalancingV2::LoadBalancer': 'Tags',
    'AWS::ElasticLoadBalancingV2::TargetGroup': 'Tags',
    # Add more resource types as they are discovered to have different tag properties
    # Default fallback is 'Tags' for most AWS resources
}

def add_default_tags(properties: Dict, schema: Dict, resource_type: str = None) -> Dict:
    """
    Add default MCP server tags to resource properties using the correct tag property name.
    
    This function automatically adds management tags for tracking resources created by the MCP server.
    It handles different AWS resource types that use different tag property names (e.g., EFS uses
    'FileSystemTags' instead of the standard 'Tags').
    
    Args:
        properties: The resource properties dictionary
        schema: The resource schema (for future validation)
        resource_type: AWS resource type (e.g., 'AWS::EFS::FileSystem', 'AWS::S3::Bucket')
    
    Returns:
        Modified properties dictionary with default tags added using the correct property name
    """
    if not properties:
        return {}
    
    # Create a copy to avoid modifying the original
    modified_properties = properties.copy()
    
    # Determine the correct tag property name for this resource type
    tag_property = RESOURCE_TAG_PROPERTY_MAP.get(resource_type, 'Tags')
    
    # Ensure the tag property exists
    if tag_property not in modified_properties:
        modified_properties[tag_property] = []
    
    # Get existing tag keys to avoid duplicates
    existing_tags = modified_properties.get(tag_property, [])
    existing_tag_keys = [tag.get('Key') for tag in existing_tags if isinstance(tag, dict)]
    
    # Define default tags that should be added to all resources
    default_tags = [
        {
            'Key': 'MANAGED_BY', 
            'Value': 'CCAPI-MCP-SERVER'
        },
        {
            'Key': 'MCP_SERVER_SOURCE_CODE', 
            'Value': 'https://github.com/awslabs/mcp/tree/main/src/ccapi-mcp-server'
        },
        {
            'Key': 'MCP_SERVER_VERSION', 
            'Value': get_package_version()
        }
    ]
    
    # Add missing default tags
    for default_tag in default_tags:
        if default_tag['Key'] not in existing_tag_keys:
            modified_properties[tag_property].append(default_tag)
    
    return modified_properties


def validate_patch(patch_document: Any) -> None:
    """
    Validate patch document for CloudControl API.
    
    Args:
        patch_document: JSON Patch document to validate
        
    Raises:
        ValueError: If patch document is invalid
    """
    if not isinstance(patch_document, list):
        raise ValueError("Patch document must be a list")
    
    valid_operations = ['add', 'remove', 'replace', 'move', 'copy', 'test']
    
    for i, operation in enumerate(patch_document):
        if not isinstance(operation, dict):
            raise ValueError(f"Patch operation {i} must be a dictionary")
        
        if 'op' not in operation:
            raise ValueError(f"Patch operation {i} missing required 'op' field")
        
        if operation['op'] not in valid_operations:
            raise ValueError(f"Invalid operation '{operation['op']}' in patch {i}")
        
        if 'path' not in operation:
            raise ValueError(f"Patch operation {i} missing required 'path' field")


def progress_event(response_event: Dict, hooks_events: List = None) -> Dict[str, Any]:
    """
    Process CloudControl API response event and format for output.
    
    Args:
        response_event: Response from CloudControl API
        hooks_events: Optional list of hook events
        
    Returns:
        Formatted progress event dictionary
    """
    if hooks_events is None:
        hooks_events = []
    
    # Extract basic information from the response
    event_data = {
        'status': response_event.get('OperationStatus', 'UNKNOWN'),
        'resource_type': response_event.get('ResourceModel', {}).get('resourceType'),
        'is_complete': response_event.get('OperationStatus') in ['SUCCESS', 'FAILED'],
        'request_token': response_event.get('RequestToken'),
        'identifier': response_event.get('ResourceModel', {}).get('identifier'),
        'event_time': response_event.get('EventTime')
    }
    
    # Add hook events if present
    if hooks_events:
        event_data['hook_events'] = hooks_events
    
    # Add error information if operation failed
    if event_data['status'] == 'FAILED':
        event_data['error_message'] = response_event.get('StatusMessage', 'Unknown error')
        event_data['error_code'] = response_event.get('ErrorCode')
    
    return event_data


def get_resource_tag_property(resource_type: str) -> str:
    """
    Get the correct tag property name for a given AWS resource type.
    
    Args:
        resource_type: AWS resource type (e.g., 'AWS::EFS::FileSystem')
        
    Returns:
        The tag property name to use ('Tags', 'FileSystemTags', etc.)
    """
    return RESOURCE_TAG_PROPERTY_MAP.get(resource_type, 'Tags')


def add_resource_to_tag_map(resource_type: str, tag_property: str) -> None:
    """
    Add a new resource type to the tag property mapping.
    
    This is useful for extending support to new resource types that use
    non-standard tag properties.
    
    Args:
        resource_type: AWS resource type (e.g., 'AWS::NewService::Resource')  
        tag_property: Tag property name (e.g., 'ResourceTags')
    """
    RESOURCE_TAG_PROPERTY_MAP[resource_type] = tag_property