namespace LitterRobot.Models.Shared;

/// <summary>
/// The shared resource across the application
/// </summary>
public record Resource
{
    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string LitterRobotId { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string LitterRobotSerial { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string PowerStatus { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string UnitStatus { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string UnitStatusText { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string LitterRobotNickname { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool Power { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool Cycle { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool PanelLockActive { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool NightLightActive { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool DidNotifyOffline { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool DFITriggered { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public bool SleepModeActive { get; init; } = false;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public string SleepMode { get; init; } = string.Empty;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long CycleCount { get; init; } = long.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long CycleCapacity { get; init; } = long.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long CyclesAfterDrawerFull { get; init; } = long.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long DFICycleCount { get; init; } = long.MinValue;

    /// <summary>
    /// 
    /// </summary>
    /// <value></value>
    public long CleanCycleWaitTimeMinutes { get; init; } = long.MinValue;

    /// <summary>
    ///  
    /// </summary>
    public const string ON_ONE = "1";

    /// <summary>
    /// 
    /// </summary>
    public const string CC = "CC";
}
