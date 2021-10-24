using System;

namespace LitterRobot.Models.Source;

/// <summary>
/// The response from the source
/// </summary>
public record Response
{
    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string PowerStatus { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string SleepModeStartTime { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public DateTimeOffset LastSeen { get; init; } = DateTime.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string SleepModeEndTime { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool AutoOfflineDisabled { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public DateTime SetupDate { get; init; } = DateTime.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long DFICycleCount { get; init; } = long.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string CleanCycleWaitTimeMinutes { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string UnitStatus { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool IsOnboarded { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string DeviceType { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string LitterRobotNickname { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long CycleCount { get; init; } = long.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string PanelLockActive { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long CyclesAfterDrawerFull { get; init; } = long.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string LitterRobotSerial { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long CycleCapacity { get; init; } = long.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string LitterRobotId { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string NightLightActive { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool DidNotifyOffline { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string IsDFITriggered { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string SleepModeActive { get; init; } = string.Empty;
}
